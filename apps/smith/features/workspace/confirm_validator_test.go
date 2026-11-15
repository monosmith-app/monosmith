package workspace

import (
	"strings"
	"testing"
)

func TestValidateConfirm(t *testing.T) {
	for value, ok := range map[string]bool{"": true, "always": true, "risky": true, "never": false, "RISKY": false} {
		cfg := &Config{}
		cfg.Workspace.Confirm = value
		err := validateConfirm(cfg)
		if (err == nil) != ok {
			t.Errorf("%q: %v", value, err)
		}
		if err != nil && !strings.Contains(err.Error(), "risky") {
			t.Errorf("the message names the choices: %v", err)
		}
	}
}
