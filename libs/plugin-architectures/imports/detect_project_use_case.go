package imports

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	goModuleLine = regexp.MustCompile(`(?m)^module[ \t]+(\S+)`)
	dartNameLine = regexp.MustCompile(`(?m)^name:[ \t]*(\S+)`)
)

// Detect reads the facts about the project in dir that its language's
// imports are read against. A missing manifest leaves the fact empty, and
// the imports that need it are then not seen.
func Detect(dir, language string) Project {
	var p Project
	if l, ok := For(language); !ok || l.Name == "" {
		return p
	}
	switch canonical(language) {
	case "go":
		if raw, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
			if m := goModuleLine.FindSubmatch(raw); m != nil {
				p.GoModule = strings.Trim(string(m[1]), `"`)
			}
		}
	case "dart":
		if raw, err := os.ReadFile(filepath.Join(dir, "pubspec.yaml")); err == nil {
			if m := dartNameLine.FindSubmatch(raw); m != nil {
				p.DartPackage = strings.Trim(string(m[1]), `"'`)
			}
		}
	case "python":
		p.PythonRoots = map[string]bool{}
		for _, base := range []string{dir, filepath.Join(dir, "src")} {
			entries, err := os.ReadDir(base)
			if err != nil {
				continue
			}
			for _, e := range entries {
				name := e.Name()
				if strings.HasPrefix(name, ".") {
					continue
				}
				if e.IsDir() {
					p.PythonRoots[name] = true
				} else if stem, ok := strings.CutSuffix(name, ".py"); ok {
					p.PythonRoots[stem] = true
				}
			}
		}
	}
	return p
}

func canonical(language string) string {
	if c, ok := aliases[language]; ok {
		return c
	}
	return language
}
