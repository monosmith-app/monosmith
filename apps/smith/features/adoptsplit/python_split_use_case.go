package adoptsplit

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
)

// SplitPython cuts the pyproject.toml of a Python project that moved into
// req.Dir. The development dependencies that go to the workspace leave its
// dev lists and join the root requirements-dev.txt (the workspace's
// toolchain, which smith creates with its own pinned tools when the root has
// none); the [tool.ruff] tables become a root ruff.toml. The runtime
// dependencies, [project], the build system and the other [tool.*] tables
// (mypy, pytest, setuptools: their paths are relative to the project) stay.
// Nothing is staged or committed.
//
// Parameters: req is the move; Hoist names the development dependencies that go.
// Returns: the notes for the person; an error, with nothing written, when
// pyproject.toml cannot be read.
func SplitPython(req Request) (Result, error) {
	var res Result
	manifest := filepath.Join(req.Root, filepath.FromSlash(req.Dir), "pyproject.toml")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return res, err
	}
	text := string(raw)

	var moved []string
	for _, g := range devGroups {
		entries, ok := readDevGroup(text, g)
		if !ok {
			continue
		}
		var remain []string
		for _, e := range entries {
			if slices.Contains(req.Hoist, RequirementName(e)) {
				moved = append(moved, e)
			} else {
				remain = append(remain, e)
			}
		}
		if len(remain) != len(entries) {
			text = writeDevGroup(text, g, remain)
		}
	}
	if len(moved) > 0 {
		added, err := addToolchainRequirements(req.Root, moved)
		if err != nil {
			return res, err
		}
		res.Notes = append(res.Notes, fmt.Sprintf("%d development dependencies left the project for the root %s (smith's own pins win for the tools it lists already); new there: %s.", len(moved), venv.RequirementsFile, strings.Join(added, ", ")))
	}

	text = dropEmptyTables(text, "project.optional-dependencies", "dependency-groups")
	rest, ruff := takeRuffTables(text)
	if ruff != "" {
		target := filepath.Join(req.Root, "ruff.toml")
		if _, err := os.Stat(target); err == nil {
			res.Notes = append(res.Notes, "the [tool.ruff] tables stay in pyproject.toml: the root already has a ruff.toml.")
		} else {
			if err := os.WriteFile(target, []byte(ruff), 0o644); err != nil {
				return res, err
			}
			text = rest
			res.Notes = append(res.Notes, "the [tool.ruff] tables are in the root ruff.toml, which ruff reads for every project; check that the paths in it (src, exclude) suit the root.")
		}
	}
	if err := os.WriteFile(manifest, []byte(text), 0o644); err != nil {
		return res, err
	}
	return res, nil
}

// addToolchainRequirements adds the requirements to the root
// requirements-dev.txt, which it creates from smith's pinned tools when the
// root has none. A requirement whose name the file lists already is not added:
// the file's own pin wins. It returns the requirements it added.
func addToolchainRequirements(root string, requirements []string) ([]string, error) {
	path := filepath.Join(root, venv.RequirementsFile)
	current, err := os.ReadFile(path)
	if err != nil {
		current = []byte(venv.Requirements)
	}
	have := map[string]bool{}
	for _, line := range strings.Split(string(current), "\n") {
		if name := RequirementName(line); name != "" && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			have[name] = true
		}
	}
	var added []string
	text := strings.TrimRight(string(current), "\r\n") + "\n"
	for _, r := range requirements {
		name := RequirementName(r)
		if have[name] {
			continue
		}
		have[name] = true
		if len(added) == 0 {
			text += "# Moved from the project's pyproject.toml when it joined the workspace.\n"
		}
		text += r + "\n"
		added = append(added, r)
	}
	if len(added) == 0 && err == nil {
		return nil, nil
	}
	return added, os.WriteFile(path, []byte(text), 0o644)
}
