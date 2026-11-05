package releasepack

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// NewBundle lists the packages of a release run and works out, for each
// ecosystem, the gh commands that download them and the command that installs
// them together. Packages and ecosystems are sorted, so the same run gives the
// same bundle.
//
// Parameters: repository is owner/name; packages are the packed packages of the run.
// Returns: the bundle.
func NewBundle(repository string, packages []Package) Bundle {
	sorted := append([]Package(nil), packages...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Ecosystem != sorted[j].Ecosystem {
			return sorted[i].Ecosystem < sorted[j].Ecosystem
		}
		return sorted[i].Project < sorted[j].Project
	})
	b := Bundle{Repository: repository, Packages: sorted, Install: []Install{}}
	for i := 0; i < len(sorted); {
		j := i
		for j < len(sorted) && sorted[j].Ecosystem == sorted[i].Ecosystem {
			j++
		}
		b.Install = append(b.Install, installOf(repository, sorted[i:j]))
		i = j
	}
	return b
}

// installOf builds the download and install commands of one ecosystem's packages.
func installOf(repository string, packages []Package) Install {
	eco := Install{Ecosystem: packages[0].Ecosystem}
	var kind Ecosystem
	for _, e := range ecosystems {
		if e.Name == eco.Ecosystem {
			kind = e
		}
	}
	var names []string
	seen := map[string]bool{}
	for _, p := range packages {
		if !seen[p.Tag] {
			seen[p.Tag] = true
			patterns := make([]string, len(kind.Patterns))
			for i, pat := range kind.Patterns {
				patterns[i] = "--pattern '" + pat + "'"
			}
			patterns = append(patterns, "--pattern '*"+ChecksumSuffix+"'")
			eco.Download = append(eco.Download, fmt.Sprintf("gh release download %q --repo %s %s --dir .", p.Tag, repository, strings.Join(patterns, " ")))
		}
		for _, f := range p.Files {
			names = append(names, f.Name)
		}
	}
	eco.Command = installCommand(kind, names)
	return eco
}

// MergeBundles adds the packages of an older bundle that the new one does not
// list (by tag and project), so a re-run that finishes some packages of a run
// does not take the others out of the bundle of their release.
func MergeBundles(older, newer Bundle) Bundle {
	have := map[string]bool{}
	for _, p := range newer.Packages {
		have[p.Tag+"\x00"+p.Project] = true
	}
	merged := append([]Package(nil), newer.Packages...)
	for _, p := range older.Packages {
		if !have[p.Tag+"\x00"+p.Project] {
			merged = append(merged, p)
		}
	}
	return NewBundle(newer.Repository, merged)
}

// BundleJSON renders release-bundle.json.
func BundleJSON(b Bundle) ([]byte, error) {
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

// BundleInstructions renders INSTALL.md: how to fetch the files of the run and
// install them together, with the checksum of each.
func BundleInstructions(b Bundle) string {
	var out strings.Builder
	out.WriteString("# Install the packages of this release\n\n")
	fmt.Fprintf(&out, "These packages were released together from %s. They are attached to GitHub Releases and are not published to a public registry. Each file keeps its normal dependency ranges, so install the packages of one ecosystem together: the files satisfy each other.\n\n", b.Repository)
	out.WriteString("A private repository needs a sign-in to download: run `gh auth login`, or set `GH_TOKEN` to a token that can read the contents of " + b.Repository + ".\n")
	for _, install := range b.Install {
		fmt.Fprintf(&out, "\n## %s\n\nDownload, in one folder:\n\n", install.Ecosystem)
		out.WriteString(codeBlock(strings.Join(install.Download, "\n")))
		out.WriteString("Check the files (each has a `.sha256` file):\n\n")
		out.WriteString(codeBlock("sha256sum -c *.sha256"))
		out.WriteString("Install:\n\n")
		out.WriteString(codeBlock(install.Command))
		out.WriteString("| Package | Version | File | SHA-256 |\n|---|---|---|---|\n")
		for _, p := range b.Packages {
			if p.Ecosystem != install.Ecosystem {
				continue
			}
			for _, f := range p.Files {
				fmt.Fprintf(&out, "| %s | %s | %s | %s |\n", p.Project, p.Version, f.Name, f.SHA256)
			}
		}
	}
	return out.String()
}

func codeBlock(text string) string { return "```sh\n" + text + "\n```\n\n" }
