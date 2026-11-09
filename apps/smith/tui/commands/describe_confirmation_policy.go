package commands

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"slices"
)

// Describe says, for the confirmation dialog, what a command that changes the
// workspace does: a title and one sentence. The dialog adds the exact command line.
func Describe(argv []string) (title, why string) {
	if len(argv) == 0 {
		return words.Say("commands.describe.default.title"), words.Say("commands.describe.default.why")
	}
	switch verb := argv[0]; {
	case verb == "release" && slices.Contains(argv, "--force"):
		return words.Say("commands.describe.release.forced.title"), words.Say("commands.describe.release.forced.why")
	case verb == "release":
		return words.Say("commands.describe.release.title"), words.Say("commands.describe.release.why")
	case verb == "doctor" && slices.Contains(argv, "--fix") && slices.Contains(argv, "--only"):
		return words.Say("commands.describe.doctor.only.title"), words.Say("commands.describe.doctor.only.why")
	case verb == "doctor" && slices.Contains(argv, "--fix"):
		return words.Say("commands.describe.doctor.title"), words.Say("commands.describe.doctor.why")
	case verb == "completion":
		return words.Say("commands.describe.completion.title"), words.Say("commands.describe.completion.why")
	case verb == "install" && slices.ContainsFunc(argv, func(a string) bool { return flagvocabulary.Is("project", a) }):
		return words.Say("commands.describe.install.packages.title"), words.Say("commands.describe.install.packages.why")
	case verb == "install":
		return words.Say("commands.describe.install.title"), words.Say("commands.describe.install.why")
	case verb == "upgrade":
		return words.Say("commands.describe.upgrade.title"), words.Say("commands.describe.upgrade.why")
	case verb == "ci":
		return words.Say("commands.describe.ci.title"), words.Say("commands.describe.ci.why")
	case verb == "add":
		return words.Say("commands.describe.add.title"), words.Say("commands.describe.add.why")
	case verb == "remove":
		return words.Say("commands.describe.remove.title"), words.Say("commands.describe.remove.why")
	case verb == "new":
		return words.Say("commands.describe.new.title"), words.Say("commands.describe.new.why")
	case verb == "config":
		return words.Say("commands.describe.config.title"), words.Say("commands.describe.config.why")
	case verb == "deps":
		if slices.Contains(argv, "--align") {
			return words.Say("commands.describe.deps.align.title"), words.Say("commands.describe.deps.align.why")
		}
		return words.Say("commands.describe.deps.title"), words.Say("commands.describe.deps.why")
	case verb == "swap":
		return words.Say("commands.describe.swap.title"), words.Say("commands.describe.swap.why")
	case verb == "bootstrap":
		return words.Say("commands.describe.bootstrap.title"), words.Say("commands.describe.bootstrap.why")
	case verb == "commit":
		return words.Say("commands.describe.commit.title"), words.Say("commands.describe.commit.why")
	case verb == "undo":
		return words.Say("commands.describe.undo.title"), words.Say("commands.describe.undo.why")
	case verb == "agent" && len(argv) > 1 && argv[1] == "file":
		return words.Say("commands.describe.agent.title"), words.Say("commands.describe.agent.why")
	case verb == "derive" && slices.Contains(argv, "--eject"):
		return words.Say("commands.describe.derive.eject.title"), words.Say("commands.describe.derive.eject.why")
	case verb == "derive" && slices.Contains(argv, "--adopt"):
		return words.Say("commands.describe.derive.adopt.title"), words.Say("commands.describe.derive.adopt.why")
	case verb == "derive" && slices.Contains(argv, "--replace"):
		return words.Say("commands.describe.derive.replace.title"), words.Say("commands.describe.derive.replace.why")
	case verb == "derive":
		return words.Say("commands.describe.derive.title"), words.Say("commands.describe.derive.why")
	case verb == "recipe" && len(argv) > 1 && argv[1] == "new":
		file := "<name>.recipe.toml"
		if i := slices.Index(argv, "--name"); i >= 0 && i+1 < len(argv) {
			file = argv[i+1] + ".recipe.toml"
		}
		return words.Say("commands.describe.recipe.title"), words.Say("commands.describe.recipe.why", "file", file)
	case verb == "plugin" && len(argv) > 1 && argv[1] == "install":
		return words.Say("commands.describe.plugin.install.title"), words.Say("commands.describe.plugin.install.why")
	case verb == "plugin":
		return words.Say("commands.describe.plugin.title"), words.Say("commands.describe.plugin.why")
	}
	return words.Say("commands.describe.other.title", "verb", argv[0]), words.Say("commands.describe.other.why")
}
