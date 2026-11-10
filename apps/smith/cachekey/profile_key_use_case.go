package cachekey

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// safeKey is true when a shell line can hold the key between double quotes
// without any escape.
func safeKey(key string) bool {
	if key == "" {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.', r == '+', r == '/', r == '=':
		default:
			return false
		}
	}
	return true
}

// WriteProfileKey puts the key in the file of a shell, between the markers.
// A block that is already there is replaced, never repeated. The rest of the
// file is not touched, and the file keeps its mode. A new file is readable by
// the user only. The caller has shown the file and asked before it calls.
//
// Parameters: env is the machine; t is the file; key is the secret.
// Returns: the error says that the key cannot be put in a shell line or that
// the file could not be read or written. No error holds the key.
func WriteProfileKey(env shellenv.Environment, t ProfileTarget, key string) error {
	if !safeKey(key) {
		return errors.New(words.Say("cachekey.profile.unsafe"))
	}
	raw, err := env.ReadFile(t.Path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errors.New(words.Say("cachekey.profile.failed", "path", t.Path, "reason", scrub(err, key)))
	}
	if err := env.WriteFilePrivate(t.Path, []byte(profileMarkers.With(string(raw), t.Line(key)))); err != nil {
		return errors.New(words.Say("cachekey.profile.failed", "path", t.Path, "reason", scrub(err, key)))
	}
	return nil
}

// RemoveProfileKeys takes the block out of every profile file that holds it.
// A file that holds nothing else is deleted.
//
// Parameters: env is the machine.
// Returns: the files that changed. The error says that a file could not be
// written; the files before it are already changed.
func RemoveProfileKeys(env shellenv.Environment) ([]string, error) {
	var changed []string
	for _, t := range ProfileTargets(env) {
		raw, err := env.ReadFile(t.Path)
		if err != nil {
			continue
		}
		rest, found := profileMarkers.Without(string(raw))
		if !found {
			continue
		}
		if strings.TrimSpace(rest) == "" {
			err = env.Remove(t.Path)
		} else {
			err = env.WriteFilePrivate(t.Path, []byte(rest))
		}
		if err != nil {
			return changed, errors.New(words.Say("cachekey.profile.failed", "path", t.Path, "reason", err.Error()))
		}
		changed = append(changed, t.Path)
	}
	return changed, nil
}

// scrub is the text of err with every copy of the secret replaced.
func scrub(err error, secret string) string {
	return strings.ReplaceAll(err.Error(), secret, MaskedKey)
}
