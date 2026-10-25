package pluginadmit

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

// Admit lets the plugins start only when each is trusted. A plugin that is
// not is offered to the person when a Prompter is set and it is not CI;
// declining, or having nobody to ask, refuses with a problem that names
// `smith plugin trust`. Nothing is started here: the caller starts the plugins
// only when Admit returns nil.
func Admit(workspace string, plugins []Plugin) error {
	if len(plugins) == 0 {
		return nil
	}
	store, err := plugintrust.OpenStore()
	if err != nil {
		return err
	}
	inCI := plugintrust.InCI(os.Getenv)
	prompter := CurrentPrompter()
	for _, st := range Describe(store, workspace, plugins) {
		if st.Trusted {
			continue
		}
		if st.Problem != "" || inCI || prompter == nil {
			return Refusal(st, inCI)
		}
		if !Ask(prompter, st) {
			return Refusal(st, inCI)
		}
		if err := store.Trust(workspace, st.Name, st.Fingerprint); err != nil {
			return fmt.Errorf("recording that you trust %s: %w", st.Name, err)
		}
		fmt.Fprintf(prompter.Out, "Trusted %q. `smith plugin untrust %s` takes it back.\n\n", st.Name, st.Name)
	}
	return nil
}
