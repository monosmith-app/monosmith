package toolrun

import "testing"

func TestStepSaysWhatIsAboutToHappen(t *testing.T) {
	log := captureLog(t)
	Step("Creating the Python environment")
	if got := log.String(); got != "Creating the Python environment ...\n" {
		t.Fatalf("%q", got)
	}
}
