package adoptsplit

// goRules: go.mod, go.sum and the code stay with the project. golangci-lint
// looks for its configuration in the folder and then upwards, so one file at
// the root serves every project. Splitting go.mod and a Go workspace file
// (go.work) is the next step: today the module moves whole.
var goRules = Rules{
	Ecosystem: Go,
	Workspace: []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"},
}

// pythonRules: the tool configuration files that ruff and flake8 find by
// walking upwards stay at the root. The [tool.*] tables inside pyproject.toml
// are not split yet: pyproject.toml moves whole.
var pythonRules = Rules{
	Ecosystem: Python,
	Workspace: []string{"ruff.toml", ".ruff.toml", ".flake8"},
}

// rustRules: rustfmt, clippy and the toolchain pin apply to the whole Cargo
// workspace and rustup finds them upwards, so they stay at the root. The
// crate's Cargo.toml moves whole: a [workspace] Cargo.toml for the root is the
// next step.
var rustRules = Rules{
	Ecosystem: Rust,
	Workspace: []string{"rustfmt.toml", ".rustfmt.toml", "clippy.toml", ".clippy.toml", "rust-toolchain", "rust-toolchain.toml"},
}
