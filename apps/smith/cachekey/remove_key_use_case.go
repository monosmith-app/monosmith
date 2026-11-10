package cachekey

import (
	"errors"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// RemoveKey undoes what the key screens did, for the key and for the SAS
// token: on Windows it unsets the user variables, elsewhere it takes the marked
// blocks out of the shell profiles. It says what changed. The person asked for
// it by name, so it does not ask again.
//
// Parameters: u is the machine; out is where the messages go.
// Errors: a variable could not be unset or a profile could not be written.
func RemoveKey(u UserEnvironment, out io.Writer) error {
	if u.CanSet() {
		for _, v := range Variables {
			if err := u.Unset(v); err != nil {
				return errors.New(words.Say("cachekey.remove.failed", "reason", err.Error()))
			}
			fmt.Fprintln(out, words.Say("cachekey.remove.windows", "variable", v))
		}
		return nil
	}
	found := false
	if u.Machine != nil {
		for _, v := range Variables {
			changed, err := RemoveProfileKeys(u.Machine, v)
			for _, path := range changed {
				found = true
				fmt.Fprintln(out, words.Say("cachekey.remove.profile", "variable", v, "path", path))
			}
			if err != nil {
				return err
			}
		}
	}
	if !found {
		_, err := fmt.Fprintln(out, words.Say("cachekey.remove.none"))
		return err
	}
	return nil
}
