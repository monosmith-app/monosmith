package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"

// pythonToolchain is the version of the first interpreter that answers, in
// the order the plugin looks for one (py -3 on Windows, python, python3).
func pythonToolchain() string {
	for _, argv := range [][]string{{"py", "-3", "--version"}, {"python", "--version"}, {"python3", "--version"}} {
		if v := toolrun.Toolchain("python", argv...); v != "" {
			return v
		}
	}
	return ""
}
