package lambda

import (
	"errors"
	"flag"
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
}

// getEnv returns env var if set, otherwise default.
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseBoolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func parseIntEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

// parseCommandLineArgs reads flags, allowing env vars as defaults.
// Flags take precedence over env variables (your tests rely on this).
func parseCommandLineArgs() Input {
	// Defaults from env
	defType := getEnv("TYPE", "config-rule-evaluation")
	defRule := getEnv("CONFIG_RULE_NAME", "")
	defRegion := getEnv("REGION", "")
	defBatch := parseIntEnv("BATCH_SIZE", 10)
	defDryRun := parseBoolEnv("DRY_RUN", false)

	typ := flag.String("type", defType, "request type (e.g. config-rule-evaluation)")
	rule := flag.String("config-rule", defRule, "AWS Config rule name")
	region := flag.String("region", defRegion, "AWS region")
	batch := flag.Int("batch-size", defBatch, "batch size for processing")
	dry := flag.Bool("dry-run", defDryRun, "simulate actions without changes")

	// NOTE: tests invoke this in isolation; ensure flags can be re-parsed in tests.
	// When tests call flag.CommandLine = flag.NewFlagSet(...) they control args.

	flag.Parse()

	return Input{
		Type:           strings.TrimSpace(*typ),
		ConfigRuleName: strings.TrimSpace(*rule),
		Region:         strings.TrimSpace(*region),
		BatchSize:      *batch,
		DryRun:         *dry,
	}
}

func validateInput(in Input) error {
	if in.Type == "" {
		return errors.New("type is required")
	}
	// keep simple allow-list; tests only require non-empty / known type
	switch in.Type {
	case "config-rule-evaluation":
		// ok
	default:
		return errors.New("invalid type")
	}

	if in.ConfigRuleName == "" {
		return errors.New("config rule name is required")
	}
	if in.Region == "" {
		return errors.New("region is required")
	}
	if in.BatchSize <= 0 {
		return errors.New("batch size must be > 0")
	}
	// put a sensible upper bound to catch mistakes; adjust if your tests expect a different cap
	if in.BatchSize > 1000 {
		return errors.New("batch size too large")
	}
	return nil
}
