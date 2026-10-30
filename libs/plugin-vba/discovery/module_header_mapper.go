package discovery

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

// Manifests are the extensions of exported VBA modules: standard modules,
// class modules and forms. The Office document itself (.xlsm, .docm) is a
// binary; the project is the modules exported from it as text.
var Manifests = []string{"*.bas", "*.cls", "*.frm"}

// MainName is the main module of a project that has one by that name.
const MainName = "main.bas"

// Header is what smith reads from the comments of a main module.
type Header struct {
	Kind, Stack, Version string
}

var (
	header = regexp.MustCompile(`(?m)^'\s*smith(Kind|Stack|Version):\s*(\S+)`)
	// vbName is the line every exported VBA module starts with, which tells
	// it from a LaTeX class or a QBasic source that shares its extension.
	vbName     = regexp.MustCompile(`(?m)^Attribute VB_Name\b[^\r\n]*(\r?\n)`)
	extensions = map[string]int{".bas": 0, ".cls": 1, ".frm": 2}
)

// ParseHeader reads the smithKind, smithStack and smithVersion comments.
func ParseHeader(src string) Header {
	var h Header
	for _, m := range header.FindAllStringSubmatch(src, -1) {
		switch m[1] {
		case "Kind":
			h.Kind = m[2]
		case "Stack":
			h.Stack = m[2]
		case "Version":
			h.Version = m[2]
		}
	}
	return h
}

// IsModule reports whether the source is an exported VBA module.
func IsModule(src string) bool { return vbName.MatchString(src) }

// FindMain is the main module of a directory, from the base names of its
// module files and a reader of them: main.bas, else the first .bas by name,
// else the first .cls, else the first .frm; the first of those that is a VBA
// module. It returns "" when the directory holds none.
func FindMain(names []string, read func(name string) (string, bool)) (name, src string) {
	var modules []string
	for _, n := range names {
		if _, ok := extensions[strings.ToLower(path.Ext(n))]; ok {
			modules = append(modules, n)
		}
	}
	sort.Slice(modules, func(i, j int) bool {
		a, b := modules[i], modules[j]
		if (a == MainName) != (b == MainName) {
			return a == MainName
		}
		if ea, eb := extensions[strings.ToLower(path.Ext(a))], extensions[strings.ToLower(path.Ext(b))]; ea != eb {
			return ea < eb
		}
		return a < b
	})
	for _, n := range modules {
		if s, ok := read(n); ok && IsModule(s) {
			return n, s
		}
	}
	return "", ""
}

// SetField writes one header comment (field is Kind, Stack or Version),
// replacing it or adding it under the module's Attribute VB_Name line (or as
// the first line of a file without one), in the file's own line ending. It
// says whether the module changed.
func SetField(src, field, value string) (string, bool) {
	line := "' smith" + field + ": " + value
	re := regexp.MustCompile(`(?m)^'\s*smith` + field + `:[^\r\n]*`)
	if re.MatchString(src) {
		out := re.ReplaceAllLiteralString(src, line)
		return out, out != src
	}
	eol := "\n"
	if strings.Contains(src, "\r\n") {
		eol = "\r\n"
	}
	if loc := vbName.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + line + eol + src[loc[1]:], true
	}
	return line + eol + src, true
}
