package adoptmove

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// The choice keys.
const (
	KeyApps     = "apps"
	KeyPackages = "packages"
	KeyLibs     = "libs"
	KeyOther    = "other"
	KeyNone     = "none"
)

var folderSegment = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Choices are the five answers to "where should it go?", in the order the
// prompts number them. The suggested one is an app folder when the project looks
// like a program, else a package folder. The person may edit the name (see
// Destination), so the labels use the derived one.
func Choices(c Candidate) []Choice {
	suggested := KeyPackages
	if c.App {
		suggested = KeyApps
	}
	out := make([]Choice, 0, 5)
	for _, key := range []string{KeyApps, KeyPackages, KeyLibs} {
		dir := key + "/" + c.Name
		out = append(out, Choice{Key: key, Dir: dir, Label: dir, Default: key == suggested})
	}
	return append(out,
		Choice{Key: KeyOther, Label: "another folder"},
		Choice{Key: KeyNone, Label: "no, leave it"},
	)
}

// DefaultChoice is the number (from 1) of the suggested choice.
func DefaultChoice(choices []Choice) int {
	for i, c := range choices {
		if c.Default {
			return i + 1
		}
	}
	return 1
}

// Destination is the folder for a choice. key is one of KeyApps, KeyPackages
// and KeyLibs (the folder is key/name), or KeyOther (name is the whole folder,
// such as tools/web). It returns a problem when the name or folder is not
// one smith can use. KeyNone has no folder: the caller handles it.
func Destination(key, name string) (string, error) {
	switch key {
	case KeyApps, KeyPackages, KeyLibs:
		if err := scaffold.ValidateName("project", name); err != nil {
			return "", problem.New(err.Error(), "type a name such as web-app")
		}
		return key + "/" + name, nil
	case KeyOther:
		return CleanFolder(name)
	}
	return "", problem.New(fmt.Sprintf("%q is not a place to move the project", key), "choose apps, packages, libs or another folder")
}

// CleanFolder checks a folder the person typed (relative to the repository
// root, forward slashes) and returns it cleaned. It refuses the repository
// root itself, an absolute path, a path that leaves the repository, a hidden
// folder (smith's and git's own live there) and odd characters.
func CleanFolder(dir string) (string, error) {
	raw := strings.TrimSpace(strings.ReplaceAll(dir, `\`, "/"))
	bad := func(why string) (string, error) {
		return "", problem.New(fmt.Sprintf("%q is not a folder smith can move the project into: %s", dir, why),
			"give a folder inside the repository, such as apps/web-app")
	}
	if raw == "" || raw == "." || raw == "./" {
		return bad("it is the repository root, where the project is already")
	}
	if strings.HasPrefix(raw, "/") || (len(raw) > 1 && raw[1] == ':') {
		return bad("it must be relative to the repository root")
	}
	clean := path.Clean(raw)
	for _, seg := range strings.Split(clean, "/") {
		switch {
		case seg == "..":
			return bad("it leaves the repository")
		case strings.HasPrefix(seg, "."):
			return bad("a folder that starts with a dot is for tools")
		case !folderSegment.MatchString(seg):
			return bad("use letters, digits, dot, hyphen and underscore")
		}
	}
	return clean, nil
}
