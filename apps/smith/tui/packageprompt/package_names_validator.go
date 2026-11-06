package packageprompt

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// ParsePackages splits what was typed into package names. A name that starts
// with a dash would reach the package manager as a flag, so it is refused and
// the person is told where the option lives instead.
func ParsePackages(text string) ([]string, error) {
	names := strings.Fields(text)
	if len(names) == 0 {
		return nil, errors.New(words.Say("packageprompt.error.none"))
	}
	for _, n := range names {
		if strings.HasPrefix(n, "-") {
			return nil, errors.New(words.Say("packageprompt.error.option", "name", strconv.Quote(n)))
		}
		if strings.ContainsFunc(n, unicode.IsControl) {
			return nil, errors.New(words.Say("packageprompt.error.control", "name", strconv.Quote(n)))
		}
	}
	return names, nil
}
