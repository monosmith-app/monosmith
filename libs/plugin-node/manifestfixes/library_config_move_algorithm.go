package manifestfixes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

var (
	// configOutputs is a tsconfig option that names a folder or file the compiler writes.
	configOutputs = regexp.MustCompile(`("(?:outDir|declarationDir|tsBuildInfoFile|outFile)"\s*:\s*)"([^"]*)"`)
	// outDirLine is the line of a tsconfig that sets outDir, with its indent.
	outDirLine = regexp.MustCompile(`(?m)^([ \t]*)"outDir"\s*:`)
	// compilerOptionsOpen is the start of the compilerOptions object.
	compilerOptionsOpen = regexp.MustCompile(`"compilerOptions"\s*:\s*\{`)
	// pathsOpen is the start of the paths object.
	pathsOpen = regexp.MustCompile(`"paths"\s*:\s*\{`)
)

// moveConfigOutputs moves the outDir, declarationDir, tsBuildInfoFile and
// outFile of a tsconfig from the old build folder to newDir. Options that
// name another folder (out-tsc/jest), the comments and the layout stay.
func moveConfigOutputs(config, old, newDir string) string {
	return configOutputs.ReplaceAllStringFunc(config, func(m string) string {
		parts := configOutputs.FindStringSubmatch(m)
		return parts[1] + `"` + followPath(parts[2], old, newDir) + `"`
	})
}

// configHasOption reports whether the tsconfig sets the compiler option.
func configHasOption(config, option string) bool {
	return regexp.MustCompile(`"` + option + `"\s*:`).MatchString(string(jsondoc.StripComments([]byte(config))))
}

// addDeclarationDir adds a declarationDir option on the line before outDir, so
// a project that references this one expects the declarations where the build
// script writes them. A tsconfig without an outDir line is returned as it is.
func addDeclarationDir(config, dir string) string {
	loc := outDirLine.FindStringSubmatchIndex(config)
	if loc == nil {
		return config
	}
	eol := "\n"
	if strings.Contains(config, "\r\n") {
		eol = "\r\n"
	}
	indent := config[loc[2]:loc[3]]
	return config[:loc[0]] + indent + `"declarationDir": "` + dir + `",` + eol + config[loc[0]:]
}

// addPathAliases adds the aliases (package name to a file relative to the
// workspace) to compilerOptions.paths of the tsconfig text, keeping comments
// and layout. A name that is mapped already to the same file stays; one
// mapped to another file is returned in conflicts and left alone. It returns
// an error when the tsconfig cannot be edited with certainty: no
// compilerOptions object, or a baseUrl other than the tsconfig's folder.
func addPathAliases(config string, aliases map[string]string) (edited string, conflicts map[string]string, err error) {
	var parsed struct {
		CompilerOptions struct {
			BaseURL string              `json:"baseUrl"`
			Paths   map[string][]string `json:"paths"`
		} `json:"compilerOptions"`
	}
	if json.Unmarshal(jsondoc.StripComments([]byte(config)), &parsed) != nil {
		return config, nil, fmt.Errorf("it is not a JSON object smith can read")
	}
	if base := strings.TrimSuffix(parsed.CompilerOptions.BaseURL, "/"); base != "" && base != "." {
		return config, nil, fmt.Errorf("its baseUrl is %q, and smith writes paths relative to the file's folder", parsed.CompilerOptions.BaseURL)
	}
	open := compilerOptionsOpen.FindStringIndex(config)
	if open == nil {
		return config, nil, fmt.Errorf("it has no compilerOptions object")
	}
	conflicts = map[string]string{}
	var missing []string
	for name, target := range aliases {
		have, ok := parsed.CompilerOptions.Paths[name]
		switch {
		case !ok:
			missing = append(missing, name)
		case len(have) != 1 || cleanDir(have[0]) != cleanDir(target):
			conflicts[name] = strings.Join(have, ", ")
		}
	}
	sort.Strings(missing)
	if len(missing) == 0 {
		return config, conflicts, nil
	}
	eol := "\n"
	if strings.Contains(config, "\r\n") {
		eol = "\r\n"
	}
	entries := func(indent string) []string {
		lines := make([]string, len(missing))
		for i, name := range missing {
			key, _ := json.Marshal(name)
			value, _ := json.Marshal([]string{aliases[name]})
			lines[i] = indent + string(key) + ": " + string(value)
		}
		return lines
	}
	if loc := pathsOpen.FindStringIndex(config[open[1]:]); loc != nil {
		at := open[1] + loc[1]
		indent := indentOfLineAt(config, open[1]+loc[0]) + "  "
		lines := strings.Join(entries(indent), ","+eol)
		rest := config[at:]
		if strings.HasPrefix(strings.TrimLeft(rest, " \t\r\n"), "}") {
			return config[:at] + eol + lines + eol + indentOfLineAt(config, open[1]+loc[0]) + rest[len(rest)-len(strings.TrimLeft(rest, " \t\r\n")):], conflicts, nil
		}
		return config[:at] + eol + lines + "," + rest, conflicts, nil
	}
	indent := firstPropertyIndent(config[open[1]:])
	lines := strings.Join(entries(indent+"  "), ","+eol)
	block := eol + indent + `"paths": {` + eol + lines + eol + indent + "}"
	rest := config[open[1]:]
	if strings.HasPrefix(strings.TrimLeft(rest, " \t\r\n"), "}") {
		return config[:open[1]] + block + rest, conflicts, nil
	}
	return config[:open[1]] + block + "," + rest, conflicts, nil
}

// indentOfLineAt is the white space that starts the line the offset is on.
func indentOfLineAt(text string, offset int) string {
	start := strings.LastIndexByte(text[:offset], '\n') + 1
	line := text[start:offset]
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// firstPropertyIndent is the indent of the first line after the "{" that
// starts with a quote, or four spaces.
func firstPropertyIndent(afterBrace string) string {
	for _, line := range strings.Split(afterBrace, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, `"`) {
			return line[:len(line)-len(trimmed)]
		}
		if strings.HasPrefix(trimmed, "}") {
			break
		}
	}
	return "    "
}
