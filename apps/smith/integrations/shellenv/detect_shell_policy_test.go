package shellenv

import "testing"

func TestDetectShell(t *testing.T) {
	for _, tc := range []struct {
		name   string
		goos   string
		parent string
		vars   map[string]string
		want   Shell
		ok     bool
	}{
		{"the parent program wins", "linux", "zsh", map[string]string{"SHELL": "/bin/bash"}, Zsh, true},
		{"a launcher is not a shell, SHELL decides", "linux", "node", map[string]string{"SHELL": "/usr/bin/fish"}, Fish, true},
		{"a login shell name", "linux", "-bash", nil, Bash, true},
		{"PowerShell outside Windows sets PSModulePath", "linux", "node", map[string]string{"PSModulePath": "/x", "SHELL": "/bin/bash"}, PowerShell, true},
		{"pwsh.exe on Windows", "windows", "pwsh.exe", nil, PowerShell, true},
		{"Windows without a hint is PowerShell", "windows", "node.exe", nil, PowerShell, true},
		{"Unix without a hint is unknown", "linux", "node", nil, "", false},
	} {
		env := &tempEnv{goos: tc.goos, parent: tc.parent, vars: tc.vars}
		if got, ok := DetectShell(env); got != tc.want || ok != tc.ok {
			t.Errorf("%s: got %q, %v", tc.name, got, ok)
		}
	}
}

func TestParseShellAcceptsProgramNames(t *testing.T) {
	for in, want := range map[string]Shell{"zsh": Zsh, "/usr/bin/fish": Fish, `C:\Program Files\PowerShell\7\pwsh.exe`: PowerShell, "PowerShell.EXE": PowerShell, "-bash": Bash} {
		if got, ok := ParseShell(in); !ok || got != want {
			t.Errorf("%q: got %q, %v", in, got, ok)
		}
	}
	if _, ok := ParseShell("cmd.exe"); ok {
		t.Error("cmd is not a supported shell")
	}
}
