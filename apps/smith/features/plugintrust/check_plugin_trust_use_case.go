package plugintrust

// Request names a plugin about to start.
type Request struct {
	// Workspace is the workspace root.
	Workspace string
	Plugin    string
	// Command is the plugin's argv.
	Command []string
	// Pin is the sha256 smith.toml gives the plugin, if any.
	Pin  string
	InCI bool
}

// Decision is the verdict with what it was reached from.
type Decision struct {
	Verdict
	// Fingerprint is the plugin's command and files as they are now.
	Fingerprint Fingerprint `json:"fingerprint"`
}

// Check measures the plugin's command and decides whether it may start.
// An error means the command could not be measured (the program is missing).
func Check(store *Store, req Request) (Decision, error) {
	fp, err := Measure(req.Command)
	if err != nil {
		return Decision{}, err
	}
	facts := Facts{Current: fp, Pin: req.Pin, InCI: req.InCI}
	if !req.InCI {
		recorded, ok, err := store.Find(req.Workspace, req.Plugin)
		if err != nil {
			return Decision{}, err
		}
		if ok {
			facts.Recorded = &recorded
		}
	}
	return Decision{Verdict: Decide(facts), Fingerprint: fp}, nil
}
