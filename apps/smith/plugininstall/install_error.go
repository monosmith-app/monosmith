package plugininstall

import (
	"errors"
	"fmt"
)

// ErrAlreadyInstalled is returned when a plugin of that name is installed
// and Request.Replace is not set.
var ErrAlreadyInstalled = errors.New("plugin already installed")

// DigestMismatchError says the archive is not the one the person or the
// release vouched for. Nothing was installed.
type DigestMismatchError struct {
	// Want is the digest that was expected and From where it came from.
	Want, From string
	// Got is the digest of what was received.
	Got string
}

func (e *DigestMismatchError) Error() string {
	return fmt.Sprintf("the archive's sha256 is %s, but %s says %s", e.Got, e.From, e.Want)
}

// NoDigestError says an archive has no digest to be checked against, so
// nothing was installed: a download is only as good as the place it came
// from, and the digest is what ties it to the author's release.
type NoDigestError struct {
	// Source is the archive and Hint how to supply the digest.
	Source, Hint string
}

func (e *NoDigestError) Error() string {
	return fmt.Sprintf("%s has no sha256 to be checked against", e.Source)
}
