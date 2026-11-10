package adoptsplit

// goRules: go.mod, go.sum and the code stay with the project. golangci-lint
// looks for its configuration in the folder and then upwards, so one file at
// the root serves every project. The root go.work that makes the module part
// of a workspace is written by adopt's workspace step, not here.
var goRules = Rules{
	Ecosystem: Go,
	Workspace: []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"},
}

// pythonRules: the tool configuration files that ruff and flake8 find by
// walking upwards stay at the root. pyproject.toml moves with the project
// and SplitPython cuts it: the development tools go to requirements-dev.txt
// and the [tool.ruff] tables to ruff.toml.
var pythonRules = Rules{
	Ecosystem: Python,
	Workspace: []string{"ruff.toml", ".ruff.toml", ".flake8"},
	Sort:      SortPython,
	Split:     SplitPython,
}

// rustRules: rustfmt, clippy and the toolchain pin apply to the whole Cargo
// workspace and rustup finds them upwards, so they stay at the root. The
// crate's Cargo.toml moves with it, and SplitRust writes the workspace around
// it.
var rustRules = Rules{
	Ecosystem: Rust,
	Workspace: []string{"rustfmt.toml", ".rustfmt.toml", "clippy.toml", ".clippy.toml", "rust-toolchain", "rust-toolchain.toml"},
	Split:     SplitRust,
}
