package discovery

import "regexp"

// sdistStaging matches the `<name>-<version>` folder that `python -m build`
// stages inside the project and removes again.
var sdistStaging = regexp.MustCompile(`^[A-Za-z0-9_.]+-\d+(\.\w+)*$`)

// IsSdistStaging reports whether a folder name is a sdist staging folder: it
// is there for a moment while a build runs, so a walk must not read it.
func IsSdistStaging(name string) bool { return sdistStaging.MatchString(name) }
