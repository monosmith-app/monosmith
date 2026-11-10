package releasing

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// sourceEntry is an exports entry that points at a TypeScript source.
var sourceEntry = regexp.MustCompile(`^\./src/(.+)\.ts$`)

// Stage makes a project's build directory (dist/<project dir>) a package npm
// can pack or publish: its package.json, with the paths that pointed into the
// build directory made relative to it, and the README and LICENSE beside it.
// A project that does not build into dist/ (plain JavaScript) is packed from
// its own directory, which Stage returns unchanged.
func Stage(workspaceRoot, root string) (string, error) {
	project := filepath.Join(workspaceRoot, filepath.FromSlash(root))
	built := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.ProjectBuildDir(root)))
	if _, err := os.Stat(built); errors.Is(err, fs.ErrNotExist) {
		return project, nil
	} else if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(project, "package.json"))
	if err != nil {
		return "", err
	}
	text := RelativeToBuild(string(raw), root)
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", err
	}
	// A package that keeps ./dist/... in its package.json while the workspace
	// builds it into dist/<project dir> (the Nx layout) is published from that
	// folder: ./dist/index.js is ./index.js there.
	if _, err := os.Stat(filepath.Join(built, "dist")); errors.Is(err, fs.ErrNotExist) {
		doc.RewriteStrings(publishedPathKeys, ownDistRelative)
	}
	// The build is the package: no files list to select it (the exclusions of
	// the list, "!**/*.map", stay), and no scripts, which would run in a
	// directory they were not written for.
	keepExclusions(doc)
	doc.Delete("scripts")
	// Inside the workspace a package is read from its sources
	// (exports: ./src/index.ts); what is published is their build.
	if exports, ok := doc.String("exports"); ok {
		if m := sourceEntry.FindStringSubmatch(exports); m != nil {
			if err := doc.SetJSON("exports", map[string]string{"types": "./" + m[1] + ".d.ts", "default": "./" + m[1] + ".js"}); err != nil {
				return "", err
			}
			doc.SetString("main", "./"+m[1]+".js")
			doc.SetString("types", "./"+m[1]+".d.ts")
		}
	}
	if err := os.WriteFile(filepath.Join(built, "package.json"), doc.Bytes(), 0o644); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(project)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		upper := strings.ToUpper(e.Name())
		goesWith := strings.HasPrefix(upper, "README") || strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "LICENCE") || upper == ".NPMRC" || isNLS(e.Name())
		if e.IsDir() || !goesWith {
			continue
		}
		content, err := os.ReadFile(filepath.Join(project, e.Name()))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(built, e.Name()), content, 0o644); err != nil {
			return "", err
		}
	}
	return built, nil
}

// publishedPathKeys are the package.json keys whose strings name built files.
var publishedPathKeys = []string{"main", "module", "browser", "types", "typings", "bin", "exports"}

// ownDistRelative makes a path into the package's own dist folder
// (./dist/index.js) relative to the folder that is packed (./index.js).
func ownDistRelative(v string) string {
	switch {
	case v == "./dist" || v == "dist":
		return "."
	case strings.HasPrefix(v, "./dist/"):
		return "./" + strings.TrimPrefix(v, "./dist/")
	case strings.HasPrefix(v, "dist/"):
		return "./" + strings.TrimPrefix(v, "dist/")
	}
	return v
}

// keepExclusions replaces the files list with one that selects everything of
// the packed folder and keeps the exclusions of the original list
// ("!**/*.tsbuildinfo"); a list without exclusions goes.
func keepExclusions(doc *jsondoc.Document) {
	var kept []string
	doc.RewriteStrings([]string{"files"}, func(v string) string {
		if strings.HasPrefix(v, "!") {
			kept = append(kept, v)
		}
		return v
	})
	doc.Delete("files")
	if len(kept) > 0 {
		_ = doc.SetJSON("files", append([]string{"**"}, kept...))
	}
}

// RelativeToBuild rewrites the paths a package.json gives into the project's
// build directory (../../dist/apps/web/main.js) as paths inside it
// (./main.js).
func RelativeToBuild(packageJSON, root string) string {
	prefix := protocol.BuildDirFromProject(root)
	out := packageJSON
	for _, p := range []string{"./" + prefix, prefix} {
		out = strings.ReplaceAll(out, `"`+p+`/`, `"./`)
		out = strings.ReplaceAll(out, `"`+p+`"`, `"."`)
		out = strings.ReplaceAll(out, ` `+p+`/`, ` ./`)
	}
	return out
}

// isNLS reports whether name is a text file of a VS Code extension's manifest
// (package.nls.json, package.nls.pt-br.json). The editor reads the %key%
// texts of package.json from them, so they must sit beside it in the package:
// without them every %key% shows as written.
func isNLS(name string) bool {
	return strings.HasPrefix(name, "package.nls.") && strings.HasSuffix(name, ".json")
}
