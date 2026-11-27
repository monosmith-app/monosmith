package workspacerecipe

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Facts are what a workspace says about the plugins its projects use.
type Facts struct {
	// Projects are the workspace's projects, in any order.
	Projects []ProjectFact
	// Layouts are the code layouts smith.toml turns on ([conventions.<name>]).
	Layouts []Layout
	// Provides lists, by plugin name, the tasks a toolchain plugin provides
	// (eslint: lint). A plugin that is missing provides nothing that is known.
	Provides map[string][]string
}

// ProjectFact is one project: its language plugin and the stack it uses.
type ProjectFact struct {
	// Language is the plugin that discovered the project (go, node, python).
	Language string
	// StackName names the stack; empty or protocol.NoStack means the language
	// plugin alone.
	StackName string
	// Stack is the stack's plugins, when StackName names one.
	Stack protocol.Stack
}

// Layout is a code layout smith.toml turns on.
type Layout struct {
	// Name is the architecture and its structure plugin (vertical-slices).
	Name string
	// Languages are the languages it checks; empty means every language.
	Languages []string
}

// Reading is the recipe a workspace describes for one language.
type Reading struct {
	// Language is the language plugin the recipe fixes.
	Language string
	// Projects is how many projects use that language.
	Projects int
	// StackName is the stack most of those projects use; empty for none.
	StackName string
	// Fixed are the levels the recipe fixes, as `smith recipe new --fixed`
	// takes them: level=plugin, or level=plugin,plugin for the levels that
	// take several (toolchain, structure, deploy). They are in the order a
	// recipe lists its levels.
	Fixed []string
}
