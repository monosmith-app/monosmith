package electronapp

// Setup is what smith reads of a project's Electron setup.
type Setup struct {
	// Tool is the setup; None when the project packs no Electron app.
	Tool Tool
	// OutDir is the folder the tool packs into, relative to the project
	// (Forge's outDir, electron-builder's directories.output), or its default
	// ("out", "dist") when the configuration does not set it. Empty when smith
	// cannot read it: a value with a macro (${os}) or a configuration in a
	// format smith does not read.
	OutDir string
	// Config is the file that sets OutDir, or would set it, relative to the
	// project: package.json, forge.config.js, electron-builder.yml.
	Config string
	// Editable says smith can set OutDir in Config safely: a JSON object, a
	// YAML file without a directories key, or a Forge configuration file with
	// one outDir string or one plain exported object.
	Editable bool
}
