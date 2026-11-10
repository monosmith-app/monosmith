package cachekey

import (
	"errors"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// RemoveKey undoes what the key screen did: on Windows it unsets the user
// variable, elsewhere it takes the marked block out of the shell profiles. It
// says what changed. The person asked for it by name, so it does not ask again.
//
// Parameters: u is the machine; out is where the messages go.
// Errors: the variable could not be unset or a profile could not be written.
func RemoveKey(u UserEnvironment, out io.Writer) error {
	if u.CanSet() {
		if err := u.Unset(); err != nil {
			return errors.New(words.Say("cachekey.remove.failed", "reason", err.Error()))
		}
		_, err := fmt.Fprintln(out, words.Say("cachekey.remove.windows"))
		return err
	}
	if u.Machine == nil {
		_, err := fmt.Fprintln(out, words.Say("cachekey.remove.none"))
		return err
	}
	changed, err := RemoveProfileKeys(u.Machine)
	for _, path := range changed {
		fmt.Fprintln(out, words.Say("cachekey.remove.profile", "path", path))
	}
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		_, err = fmt.Fprintln(out, words.Say("cachekey.remove.none"))
	}
	return err
}
