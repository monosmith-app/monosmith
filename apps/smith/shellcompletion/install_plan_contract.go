package shellcompletion

import "github.com/monosmith-app/monosmith/apps/smith/shellenv"

// Target is one file that install writes.
type Target struct {
	// Path is the file.
	Path string
	// Whole is true when the file is smith's alone (a completion script). It is
	// false when the block goes into a file of the person (a profile, .zshrc,
	// .bashrc).
	Whole bool
	// Body is what goes between the markers.
	Body string
	// State is how the file stands now. PlanFor fills it.
	State State
}

// Plan is everything install would do for one shell.
type Plan struct {
	// Shell is the shell the plan is for.
	Shell shellenv.Shell
	// Targets are the files to write, in order.
	Targets []Target
	// PackagedAt is the file that the package of smith already ships for the
	// shell; "" when none does. Install changes nothing then.
	PackagedAt string
}

// State says how one target stands.
type State int

// The states of a target.
const (
	// Missing means the file or the block is not there.
	Missing State = iota
	// Current means the block is there and holds the script of this smith.
	Current
	// Outdated means the block is there and holds another script.
	Outdated
	// Foreign means a whole file is there that smith did not write.
	Foreign
)

// Change is what install or uninstall did to one file.
type Change struct {
	// Path is the file.
	Path string
	// Action is what happened: "wrote", "updated", "unchanged" or "removed".
	Action string
}
