package catalog

import "testing"

// TestTerminalNeedsAReason keeps the principle that nothing an interface
// tells the person to run needs a terminal without a stated reason.
func TestTerminalNeedsAReason(t *testing.T) {
	listed := 0
	for _, a := range All() {
		if a.Safety == NeedsTerminal {
			listed++
			if a.Terminal == "" {
				t.Errorf("%s needs a terminal but states no reason: run it in a dialog, or say why it cannot", a.ID())
			}
		}
		for _, v := range a.Variants {
			if v.Safety == NeedsTerminal && v.Terminal == "" {
				t.Errorf("%s %v needs a terminal but states no reason", a.ID(), v.When)
			}
		}
	}
	if listed == 0 {
		t.Fatal("the catalogue lists no terminal action: the check proves nothing")
	}
}
