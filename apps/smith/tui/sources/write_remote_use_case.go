package sources

import (
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// writeRemote writes `[cache] remote` the way the CLI does: through the shared
// operation log. It commits only when the person ticked the box, or when
// commit_operations is on and the work tree was clean; smith never commits
// unasked.
//
// Parameters: root is the workspace root; remote is the address; commit is the
// ticked box.
// Returns: a sentence for the person about the commit (empty when none was
// asked and none made), and the error of the write.
func writeRemote(root, remote string, commit bool) (string, error) {
	begin := operation.Begin
	if commit {
		begin = operation.BeginRequested
	}
	pending, note := begin(root)
	if err := configedit.Set(root, "cache.remote", []string{remote}); err != nil {
		return "", err
	}
	if pending == nil {
		return note, nil
	}
	hash, err := pending.Finish(words.Say("tui.cache_setup.operation"))
	switch {
	case err != nil:
		return words.Say("tui.cache_setup.not_committed", "err", err.Error()), nil
	case hash != "":
		return words.Say("tui.cache_setup.committed"), nil
	}
	return note, nil
}
