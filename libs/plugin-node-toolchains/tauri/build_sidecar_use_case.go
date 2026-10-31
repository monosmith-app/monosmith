package tauri

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// BuildSidecar builds a Rust package in release mode, for the machine's own
// platform, and returns the path of the program the sidecar names: its Bin,
// else the only program the package builds, else the one named like the
// package or the project. Cargo is incremental, so a program that is up to
// date costs a moment.
func BuildSidecar(workspaceRoot, crate string, s Sidecar) (string, error) {
	cmd := exec.Command("cargo", "build", "--release", "-p", crate, "--message-format=json-render-diagnostics")
	cmd.Dir = workspaceRoot
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("sidecar %s: cargo build --release -p %s failed: %w", s.Name, crate, err)
	}
	programs := builtPrograms(out.String())
	if len(programs) == 0 {
		return "", fmt.Errorf("sidecar %s: %s builds no program (it needs src/main.rs or a [[bin]])", s.Name, crate)
	}
	return pickProgram(programs, s, crate)
}

func pickProgram(programs []string, s Sidecar, crate string) (string, error) {
	stem := func(p string) string { return strings.TrimSuffix(filepath.Base(p), ".exe") }
	var names []string
	for _, p := range programs {
		names = append(names, stem(p))
		if s.Bin != "" && stem(p) == s.Bin {
			return p, nil
		}
	}
	if s.Bin != "" {
		return "", fmt.Errorf("sidecar %s: %s builds no program named %s (it builds %s)", s.Name, crate, s.Bin, strings.Join(names, ", "))
	}
	if len(programs) == 1 {
		return programs[0], nil
	}
	for _, want := range []string{s.Name, crate, s.Project} {
		if i := slices.Index(names, want); i >= 0 {
			return programs[i], nil
		}
	}
	return "", fmt.Errorf("sidecar %s: %s builds %s; name one with bin", s.Name, crate, strings.Join(names, ", "))
}

// builtPrograms reads cargo's JSON messages and returns the executables it built.
func builtPrograms(messages string) []string {
	var programs []string
	sc := bufio.NewScanner(strings.NewReader(messages))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		var m struct {
			Reason     string `json:"reason"`
			Executable string `json:"executable"`
		}
		if json.Unmarshal(sc.Bytes(), &m) == nil && m.Reason == "compiler-artifact" && m.Executable != "" && !slices.Contains(programs, m.Executable) {
			programs = append(programs, m.Executable)
		}
	}
	return programs
}
