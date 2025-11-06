package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"
)

type Input struct {
	Type           string
	ConfigRuleName string
	Region         string
	BatchSize      int
	DryRun         bool

	// Deprecated: compat with older tests.
	ConfigRule string
}

// envBySuffix returns the first env value whose KEY ends with any of the tokens.
// Example: with tokens ["TYPE", "REQUEST_TYPE"], it will match TYPE, REQUEST_TYPE,
// LOGGUARDIAN_TYPE, LG_REQUEST_TYPE, etc.
func envBySuffix(tokens ...string) string {
	// Precompute upper tokens with leading underscore variants too.
	alts := make([]string, 0, len(tokens)*2)
	for _, t := range tokens {
		u := strings.ToUpper(t)
		alts = append(alts, u, "_"+u)
	}
	for _, kv := range os.Environ() {
		// split only on first '='
		i := strings.IndexByte(kv, '=')
		if i <= 0 {
			continue
		}
		key := strings.ToUpper(kv[:i])
		val := kv[i+1:]
		for _, suf := range alts {
			if strings.HasSuffix(key, suf) {
				return val
			}
		}
	}
	return ""
}

// parseCommandLineArgs reads flags from a fresh FlagSet.
// 1) start with zero-values
// 2) parse CLI flags (if any)
// 3) fill missing from env by suffix (so namespaced keys work)
// Flags still override env.
func parseCommandLineArgs() Input {
	fs := flag.NewFlagSet("lambda", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	typ := fs.String("type", "", "request type (e.g. config-rule-evaluation)")
	rule := fs.String("config-rule", "", "AWS Config rule name")
	region := fs.String("region", "", "AWS region")
	batch := fs.Int("batch-size", 0, "batch size for processing")
	dry := fs.Bool("dry-run", false, "simulate actions without changes")

	_ = fs.Parse(os.Args[1:])

	cfg := Input{
		Type:           strings.TrimSpace(*typ),
		ConfigRuleName: strings.TrimSpace(*rule),
		Region:         strings.TrimSpace(*region),
		BatchSize:      *batch,
		DryRun:         *dry,
	}

	// Hydrate from env only if still empty/zero.
	if cfg.Type == "" {
		cfg.Type = strings.TrimSpace(envBySuffix("TYPE", "REQUEST_TYPE"))
	}
	if cfg.ConfigRuleName == "" {
		cfg.ConfigRuleName = strings.TrimSpace(envBySuffix("CONFIG_RULE_NAME", "CONFIG_RULE", "RULE"))
	}
	if cfg.Region == "" {
		cfg.Region = strings.TrimSpace(envBySuffix("REGION", "AWS_REGION", "AWS_DEFAULT_REGION"))
	}
	if cfg.BatchSize == 0 {
		if s := envBySuffix("BATCH_SIZE", "BATCHSIZE"); s != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
				cfg.BatchSize = n
			}
		}
	}
	if !cfg.DryRun {
		if s := envBySuffix("DRY_RUN", "DRYRUN"); s != "" {
			if b, err := strconv.ParseBool(strings.TrimSpace(s)); err == nil {
				cfg.DryRun = b
			}
		}
	}

	// back-compat mirror
	cfg.ConfigRule = cfg.ConfigRuleName
	return cfg
}

func validateInput(in Input) error {
	if in.Type == "" {
		return errors.New("type is required")
	}
	switch in.Type {
	case "config-rule-evaluation":
	default:
		return errors.New("invalid type")
	}

	rule := in.ConfigRuleName
	if rule == "" {
		rule = in.ConfigRule
	}
	if rule == "" {
		return errors.New("config rule name is required")
	}
	if in.Region == "" {
		return errors.New("region is required")
	}
	if in.BatchSize <= 0 {
		return errors.New("batch size must be > 0")
	}
	if in.BatchSize > 1000 {
		return errors.New("batch size too large")
	}
	return nil
}

// version shim for tests
var version = "dev" // can be overridden with -ldflags "-X main.version=VALUE"

func getVersion() string {
	if v := os.Getenv("VERSION"); v != "" {
		return v
	}
	return version
}
