package plugintrust

import (
	"fmt"
	"slices"
	"strings"
)

// Facts are what Decide weighs.
type Facts struct {
	// Recorded is the user's trust decision for this workspace and plugin,
	// nil when there is none.
	Recorded *Record
	// Current is what the plugin's command is made of now.
	Current Fingerprint
	// Pin is the sha256 smith.toml gives the plugin, empty when none.
	Pin string
	// InCI is true when the environment says smith runs in a CI service.
	InCI bool
}

// Decide says whether a plugin may start. On a person's machine only their
// own recorded decision counts: a pin in smith.toml came with the repository,
// so it cannot vouch for itself. In CI there is no such decision on the
// machine, so a plugin runs only when smith.toml pins the digest of its file
// and the file still has it.
func Decide(f Facts) Verdict {
	if f.InCI {
		return decideInCI(f)
	}
	r := f.Recorded
	switch {
	case r == nil:
		return Verdict{Reason: ReasonNew, Detail: "you have not trusted it in this workspace"}
	case !slices.Equal(r.Command, f.Current.Command):
		return Verdict{Reason: ReasonCommandChanged, Detail: "its command is not the one you trusted"}
	}
	if changed, ok := changedFile(r.Executables, f.Current.Executables); ok {
		return Verdict{Reason: ReasonFileChanged, Detail: changed}
	}
	return Verdict{Trusted: true, Via: "recorded"}
}

func decideInCI(f Facts) Verdict {
	pin := NormalizeDigest(f.Pin)
	if pin == "" {
		return Verdict{Reason: ReasonNotPinned, Detail: "smith.toml does not pin its sha256"}
	}
	if got := PinnedDigest(f.Current); got != pin {
		return Verdict{Reason: ReasonPinMismatch, Detail: fmt.Sprintf("smith.toml pins %s but the file is %s", short(pin), short(got))}
	}
	return Verdict{Trusted: true, Via: "pinned"}
}

// PinnedDigest is the digest `sha256` in smith.toml is compared with: the
// plugin's own file, which is the script when the command runs an
// interpreter on one, else the program.
func PinnedDigest(fp Fingerprint) string {
	if len(fp.Executables) == 0 {
		return ""
	}
	return fp.Executables[len(fp.Executables)-1].SHA256
}

// NormalizeDigest lower-cases a digest and drops spaces and a "sha256:"
// prefix, so a pin may be written either way.
func NormalizeDigest(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.TrimPrefix(s, "sha256:")
}

func changedFile(recorded, current []Executable) (string, bool) {
	if len(recorded) != len(current) {
		return "the files its command runs are not the ones you trusted", true
	}
	for i := range recorded {
		switch {
		case recorded[i].Path != current[i].Path:
			return fmt.Sprintf("it now runs %s instead of %s", current[i].Path, recorded[i].Path), true
		case recorded[i].SHA256 != current[i].SHA256:
			return fmt.Sprintf("%s: sha256 was %s, now %s", current[i].Path, short(recorded[i].SHA256), short(current[i].SHA256)), true
		}
	}
	return "", false
}

func short(digest string) string {
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}
