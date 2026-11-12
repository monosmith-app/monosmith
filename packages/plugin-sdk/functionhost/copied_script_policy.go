package functionhost

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// CopiedScripts are the project files that smith scaffolded into function and
// Electron apps before the apps ran smith's commands (`smith func package`,
// `smith electron assemble`). A project holds no copy of a smith script now.
var CopiedScripts = []string{"scripts/assemble.mjs", "scripts/package.mjs", "scripts/start.mjs"}

// CopiedPythonScripts are the files that smith scaffolded into Python function
// apps, at the project root.
var CopiedPythonScripts = []string{"package.py", "start.py"}

// quoted matches a single-quoted JavaScript string.
var quoted = regexp.MustCompile(`'[^']*'`)

// Fingerprint is the shape of a script: its lines without blanks and
// indentation, with the text of every single-quoted string left out. Two
// copies that differ only in the paths and names that were filled in at
// scaffold time have the same fingerprint, a copy with an edited line has not.
//
// Parameters: text is the script.
// Returns the hash of the shape, in hexadecimal.
func Fingerprint(text string) string { return shapeHash(text, quoted) }

// shapeHash hashes the lines of a script without blanks and indentation, with
// the strings that literal matches left out.
func shapeHash(text string, literal *regexp.Regexp) string {
	var shape []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			shape = append(shape, literal.ReplaceAllString(line, "''"))
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(shape, "\n")))
	return hex.EncodeToString(sum[:])
}

// pyQuoted matches a double-quoted Python string.
var pyQuoted = regexp.MustCompile(`"[^"]*"`)

// PythonFingerprint is Fingerprint for a Python script: the text of every
// double-quoted string is left out.
//
// Parameters: text is the script.
// Returns the hash of the shape, in hexadecimal.
func PythonFingerprint(text string) string {
	return shapeHash(text, pyQuoted)
}

// IsScaffolded reports whether a copied script is still as smith wrote it, so
// that removing it loses nothing of the person's.
//
// Parameters: text is the content of the copy.
// Returns true when its fingerprint is one of the scaffolded scripts.
func IsScaffolded(text string) bool {
	return scaffoldedPrints[Fingerprint(text)] || scaffoldedPrints[PythonFingerprint(text)]
}

// scaffoldedPrints are the fingerprints of every script smith scaffolded.
var scaffoldedPrints = map[string]bool{
	// functions: assemble, package
	"e27233cea115900d2707cea63911c7da89e91c2a1c4c9f366fb56be367fb6573": true,
	"8a84a1d591c8b3e98c5ce767525b19552e381b3b317032f25cb1d7f77439000f": true,
	// functions: start of Azure, of AWS (http, other), of GCP (http, other)
	"1cf6287102f47785d2625e6bdacc1270cdb657c7d50fb7002f50d9d447e08e76": true,
	"26ee3a9c4a94af85e64e56f3d179cfaca43bf215abc98a4cdab0e9019864f3b6": true,
	"14ef1f2e1badf5333700affdd3a1512e1d83e576ac5c393822f5ca9517269ad0": true,
	"1e64abfe4435e10c06a17dcd0d9bec6af6abfe19e2cc5e5a4e659faa3b164bb9": true,
	"3a256f1494985e65dec55fb10f258acc834cff5f271283bc5200b3eece06c472": true,
	// electron: assemble, start, package
	"118061cce4be63410a8b3c8434aaf888ab74b05b5a35144eb82d712f19be681b": true,
	"e79cd964ac401bffa87fe5e34741b78f5a27aa8ebba1afd858cdcaa9f30be833": true,
	"789170de8e858eb9183f24cabe993c33093a03364deb874d0f63b76eeeb67201": true,
	// python functions: package.py of Azure, AWS, GCP
	"4c67f78ad61edcfd34c61970d24d77fe3528a41d9776030757ce8a9d5d3acfc7": true,
	"89003c354ce9d389c57ddef2436d2f6adea1ff82153130d6723164a7bf0183b3": true,
	"629117df4109f7822943826255ff79c2cc2260742bc2900cdba941618239d126": true,
	// python functions: start.py of Azure, AWS (http, other), GCP (http, other)
	"111b5c3fef1b8d1a85ce8dbb278d1d0540372f3eba66867ac1d7a0e70bbd8aa7": true,
	"424678a74a7503d6613e58b7919f908acf5bbede76422d616102bfa92e344ce0": true,
	"ed2dc4039fb9092888b230d86848543e23ddfa4dfc9f4c02c0e03f70794d90e0": true,
	"d81effa241f2b36fbc6854865752da5431692299b77f03901ed74f256b84e622": true,
	"59e9ee0d1107f4cd5467fdfbff02ec9824fc7b79ef0f46aebec804a3885ebd4b": true,
}
