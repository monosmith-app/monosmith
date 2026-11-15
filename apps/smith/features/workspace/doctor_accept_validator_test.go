package workspace

import (
	"strings"
	"testing"
)

func TestAcceptedWarningsAreRead(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\n[[doctor.accept]]\ncheck = \"c\"\nproject = \"web\"\nreason = \"why\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Doctor.Accept) != 1 || cfg.Doctor.Accept[0].Project != "web" || cfg.Doctor.Accept[0].Reason != "why" {
		t.Errorf("entry not read: %+v", cfg.Doctor)
	}
}

func TestAnAcceptedWarningNeedsACheckAndAReason(t *testing.T) {
	for name, body := range map[string]string{
		"no reason": "check = \"c\"\n",
		"no check":  "reason = \"r\"\n",
		"blank":     "check = \"c\"\nreason = \"  \"\n",
	} {
		_, err := Parse([]byte("[workspace]\nname = \"w\"\n[[doctor.accept]]\n" + body))
		if err == nil || !strings.Contains(err.Error(), "needs a check and a reason") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
