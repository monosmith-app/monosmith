package workspace

import (
	"strings"
	"testing"
)

func TestCICostKeysAreCheckedWithAClearError(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"native_on unknown":  {"native_on = \"never\"\n", "native_on must be"},
		"audit unknown":      {"audit = \"hourly\"\n", "audit must be"},
		"triggers empty":     {"triggers = []\n", "triggers must not be empty"},
		"triggers unknown":   {"triggers = [\"push\", \"tag\"]\n", "unknown trigger \"tag\""},
		"triggers twice":     {"triggers = [\"push\", \"push\"]\n", "lists \"push\" twice"},
		"cron too short":     {"audit_cron = \"* * *\"\n", "five fields"},
		"cron holds a quote": {"audit_cron = \"0 6 * * \\\"1\"\n", "quotes"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte("[workspace]\nname = \"w\"\n[ci]\n" + tc.toml))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("want an error with %q, got %v", tc.want, err)
			}
		})
	}
}

func TestCICostKeysTakeTheirValues(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\n[ci]\nnative_on = \"main\"\ntriggers = [\"pull_request\"]\naudit = \"schedule\"\naudit_cron = \"0 3 * * 0\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	ci := cfg.CI
	if ci.NativeOn != "main" || len(ci.Triggers) != 1 || ci.Triggers[0] != "pull_request" || ci.Audit != "schedule" || ci.AuditCron != "0 3 * * 0" {
		t.Errorf("got %+v", ci)
	}
	if cfg, err = Parse([]byte("[workspace]\nname = \"w\"\n")); err != nil || cfg.CI.NativeOn != "" || cfg.CI.Triggers != nil {
		t.Errorf("defaults: %v %+v", err, cfg.CI)
	}
}
