package toolcatalog

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// lookups are the commands that ask a package manager whether it knows a
// package, by manager. A manager the machine lacks is skipped: the weekly
// workflow runs this test on Windows, macOS and Ubuntu, so each manager is
// asked where it lives.
var lookups = map[string]func(pkg string) []string{
	"winget": func(pkg string) []string {
		return []string{"winget", "show", "--id", pkg, "--exact", "--accept-source-agreements"}
	},
	"brew":      func(pkg string) []string { return []string{"brew", "info", "--formula", pkg} },
	"brew-cask": func(pkg string) []string { return []string{"brew", "info", "--cask", pkg} },
	"apt":       func(pkg string) []string { return []string{"apt-cache", "show", pkg} },
}

// TestInstallIDsExistInTheirCatalogue asks each package manager on this machine
// whether it knows the package id each install entry names, so a renamed or
// removed package is found by the weekly run and not by a user. It needs the
// network and the managers, so it runs only when SMITH_CATALOGUE_CHECK is set.
// An id with a {major} placeholder is skipped: it has no fixed name to ask for.
func TestInstallIDsExistInTheirCatalogue(t *testing.T) {
	if os.Getenv("SMITH_CATALOGUE_CHECK") == "" {
		t.Skip("set SMITH_CATALOGUE_CHECK=1 to ask the package managers")
	}
	for _, e := range entries {
		for _, in := range e.Install {
			lookup, ok := lookups[in.Manager]
			if !ok || strings.Contains(in.Package, "{") {
				continue
			}
			args := lookup(in.Package)
			if _, err := exec.LookPath(args[0]); err != nil {
				continue
			}
			t.Run(e.Name+"/"+in.Manager+"/"+in.Package, func(t *testing.T) {
				if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
					t.Errorf("%s does not know %q (%s): %s", in.Manager, in.Package, strings.Join(args, " "), strings.TrimSpace(string(out)))
				}
			})
		}
	}
}
