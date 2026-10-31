package lineendings

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// siblingSample is how many neighbouring files the policy reads to learn the
// ending a folder uses.
const siblingSample = 8

// NewFileCRLF decides whether a file smith creates at rel (slash-separated,
// from the workspace root) should use CRLF. The order is: an eol rule in the
// root git attributes file; git's core.eol; core.autocrlf (true means CRLF,
// input means LF); then the majority ending of the files in the same folder
// and in the workspace root. With none of these it is false: LF.
func NewFileCRLF(root, rel string) bool {
	if attrs, err := os.ReadFile(filepath.Join(root, ".gitattributes")); err == nil {
		switch EOLFromAttributes(string(attrs), rel) {
		case "crlf":
			return true
		case "lf":
			return false
		}
	}
	switch gitSetting(root, "core.eol") {
	case "crlf":
		return true
	case "lf":
		return false
	}
	switch gitSetting(root, "core.autocrlf") {
	case "true":
		return true
	case "input":
		return false
	}
	return siblingsUseCRLF(root, path.Dir(rel))
}

// siblingsUseCRLF reports whether most of the sampled text files in dir, then
// in the workspace root, end their lines with CRLF.
func siblingsUseCRLF(root, dir string) bool {
	crlf, lf := 0, 0
	for _, d := range []string{dir, "."} {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(d)))
		if err != nil {
			continue
		}
		seen := 0
		for _, e := range entries {
			if e.IsDir() || seen >= siblingSample || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(d), e.Name()))
			if err != nil {
				continue
			}
			if isCRLF, known := IsCRLF(string(raw)); known {
				seen++
				if isCRLF {
					crlf++
				} else {
					lf++
				}
			}
		}
		if crlf+lf > 0 {
			break
		}
	}
	return crlf > lf
}
