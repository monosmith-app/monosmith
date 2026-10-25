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
