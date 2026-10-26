package rules

// modularMonolith is one deployable split into modules under modules/, each
// reached only through its public face.
var modularMonolith = Ruleset{
	Name:    "modular-monolith",
	Modules: true,
	Summary: "a module is reached only through the files directly in its folder or its api folder, never into its internals",
}
