package workspace

import (
	"strings"
	"testing"
)

func TestTheArtifactSettingsOfCIAreRead(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\n[ci]\nartifacts = \"optional\"\nartifact_retention_days = 7\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CI.Artifacts != "optional" || cfg.CI.ArtifactRetentionDays != 7 {
		t.Errorf("settings not read: %+v", cfg.CI)
	}
}

func TestTheArtifactSettingsOfCIAreChecked(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"unknown value": {"artifacts = \"sometimes\"\n", "artifacts must be"},
		"negative":      {"artifact_retention_days = -1\n", "artifact_retention_days must be"},
		"over 90":       {"artifact_retention_days = 91\n", "artifact_retention_days must be"},
	} {
		_, err := Parse([]byte("[workspace]\nname = \"w\"\n[ci]\n" + tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}
