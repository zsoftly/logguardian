// cmd/lambda/main_test.go
package main

import (
	"os"
	"testing"
)

func TestParseCommandLineArgs(t *testing.T) {
	t.Run("default_values", func(t *testing.T) {
		os.Args = []string{"logguardian-lambda"}
		cfg := parseCommandLineArgs()
		if cfg.Type != "" || cfg.ConfigRule != "" || cfg.Region != "" || cfg.BatchSize != 0 || cfg.DryRun {
			t.Fatalf("expected zero-values, got %+v", cfg)
		}
	})

	t.Run("with_command_line_args", func(t *testing.T) {
		os.Args = []string{"logguardian-lambda",
			"-type", "config-rule-evaluation",
			"-config-rule", "cloudwatch-log-group-encrypted",
			"-region", "us-east-1",
			"-batch-size", "10",
			"-dry-run",
		}
		cfg := parseCommandLineArgs()
		if cfg.Type != "config-rule-evaluation" ||
			cfg.ConfigRule != "cloudwatch-log-group-encrypted" ||
			cfg.Region != "us-east-1" ||
			cfg.BatchSize != 10 ||
			!cfg.DryRun {
			t.Fatalf("unexpected cfg: %+v", cfg)
		}
	})

	t.Run("with_environment_variables", func(t *testing.T) {
		t.Setenv("LG_TYPE", "config-rule-evaluation")
		t.Setenv("LG_CONFIG_RULE", "r1")
		t.Setenv("LG_REGION", "eu-west-1")
		t.Setenv("LG_BATCH_SIZE", "5")
		t.Setenv("LG_DRY_RUN", "true")

		os.Args = []string{"logguardian-lambda"}
		cfg := parseCommandLineArgs()
		if cfg.Type != "config-rule-evaluation" || cfg.ConfigRule != "r1" ||
			cfg.Region != "eu-west-1" || cfg.BatchSize != 5 || !cfg.DryRun {
			t.Fatalf("unexpected cfg from env: %+v", cfg)
		}
	})

	t.Run("command_line_overrides_environment", func(t *testing.T) {
		t.Setenv("LG_TYPE", "x")
		t.Setenv("LG_CONFIG_RULE", "x")
		t.Setenv("LG_REGION", "x")
		t.Setenv("LG_BATCH_SIZE", "1")
		t.Setenv("LG_DRY_RUN", "false")

		os.Args = []string{"logguardian-lambda",
			"-type", "config-rule-evaluation",
			"-config-rule", "r2",
			"-region", "ap-south-1",
			"-batch-size", "7",
			"-dry-run",
		}
		cfg := parseCommandLineArgs()
		if cfg.Type != "config-rule-evaluation" || cfg.ConfigRule != "r2" ||
			cfg.Region != "ap-south-1" || cfg.BatchSize != 7 || !cfg.DryRun {
			t.Fatalf("override failed: %+v", cfg)
		}
	})
}

func TestValidateInput(t *testing.T) {
	t.Run("valid_input", func(t *testing.T) {
		cfg := Input{
			Type:       "config-rule-evaluation",
			ConfigRule: "cloudwatch-log-group-encrypted",
			Region:     "us-east-1",
			BatchSize:  10,
		}
		if err := validateInput(cfg); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("invalid_type", func(t *testing.T) {
		cfg := Input{Type: "bad"}
		if err := validateInput(cfg); err == nil {
			t.Fatal("expected error for invalid type")
		}
	})

	t.Run("missing_config_rule_name", func(t *testing.T) {
		cfg := Input{Type: "config-rule-evaluation", Region: "us-east-1", BatchSize: 10}
		if err := validateInput(cfg); err == nil {
			t.Fatal("expected error for missing config rule")
		}
	})

	t.Run("missing_region", func(t *testing.T) {
		cfg := Input{Type: "config-rule-evaluation", ConfigRule: "x", BatchSize: 10}
		if err := validateInput(cfg); err == nil {
			t.Fatal("expected error for missing region")
		}
	})

	t.Run("invalid_batch_size - zero", func(t *testing.T) {
		cfg := Input{Type: "config-rule-evaluation", ConfigRule: "x", Region: "us-east-1", BatchSize: 0}
		if err := validateInput(cfg); err == nil {
			t.Fatal("expected error for zero batch size")
		}
	})

	t.Run("invalid_batch_size - too_large", func(t *testing.T) {
		cfg := Input{Type: "config-rule-evaluation", ConfigRule: "x", Region: "us-east-1", BatchSize: 10001}
		if err := validateInput(cfg); err == nil {
			t.Fatal("expected error for too large batch size")
		}
	})
}

func TestGetVersion(t *testing.T) {
	t.Run("with_version_env_var", func(t *testing.T) {
		t.Setenv("VERSION", "1.2.3")
		if v := getVersion(); v != "1.2.3" {
			t.Fatalf("expected 1.2.3, got %s", v)
		}
	})
	t.Run("without_version_env_var", func(t *testing.T) {
		os.Unsetenv("VERSION")
		if v := getVersion(); v == "" {
			t.Fatal("expected non-empty default/version")
		}
	})
}

func TestGetExecutionMode(t *testing.T) {
	if !getExecutionMode(true) {
		t.Fatal("expected dry-run mode when dryRun=true")
	}
	if getExecutionMode(false) {
		t.Fatal("expected apply mode when dryRun=false")
	}
}
