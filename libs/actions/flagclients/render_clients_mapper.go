package flagclients

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

// header tells a reader that the file is written by the generator.
const header = "Generated from libs/actions/flagvocabulary/flag_vocabulary.json by `go run ./flaggen` in libs/actions. Do not edit: change the vocabulary and run it again."

// TypeScript renders the flag constants of the VS Code extension: for each
// concept the spelling a command line should use (`flag.project` is "-p") and the
// long form (`longFlag.project` is "--project"). The result is the text of
// apps/vscode/src/smith-flags/smith-flags.config.ts.
func TypeScript(concepts []flagvocabulary.Concept) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// %s\n\n", header)
	b.WriteString("/** The spelling of each shared flag a smith command line should use: the shorthand when there is one. */\n")
	b.WriteString("export const flag = {\n")
	for _, c := range concepts {
		fmt.Fprintf(&b, "  %-*s '%s',\n", keyWidth(concepts), camel(c.ID)+":", spelling(c))
	}
	b.WriteString("} as const\n\n")
	b.WriteString("/** The long form of each shared flag. */\n")
	b.WriteString("export const longFlag = {\n")
	for _, c := range concepts {
		fmt.Fprintf(&b, "  %-*s '--%s',\n", keyWidth(concepts), camel(c.ID)+":", c.Flag)
	}
	b.WriteString("} as const\n")
	return b.String()
}

// Kotlin renders the flag constants of the JetBrains plugin as the object
// SmithFlagsConfig: PROJECT is "-p", PROJECT_LONG is "--project". The result is the
// text of SmithFlagsConfig.kt in the smithcommand package.
func Kotlin(concepts []flagvocabulary.Concept) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package dev.monosmith.jetbrains.smithcommand\n\n/**\n * %s\n */\nobject SmithFlagsConfig {\n", header)
	for _, c := range concepts {
		name := strings.ToUpper(strings.ReplaceAll(c.ID, "-", "_"))
		fmt.Fprintf(&b, "    /** %s. */\n    const val %s = \"%s\"\n\n", capitalise(c.Meaning), name, spelling(c))
		fmt.Fprintf(&b, "    /** The long form of [%s]. */\n    const val %s_LONG = \"--%s\"\n\n", name, name, c.Flag)
	}
	return strings.TrimRight(b.String(), "\n") + "\n}\n"
}

// Markdown renders the reference table of the flag vocabulary, the text of
// docs/reference/flag-vocabulary.md.
func Markdown(concepts []flagvocabulary.Concept, exceptions []flagvocabulary.Exception) string {
	var b strings.Builder
	b.WriteString("# Flag vocabulary\n\n")
	fmt.Fprintf(&b, "<!-- %s -->\n\n", header)
	b.WriteString("One idea has one flag in every smith command. This page lists the ideas that several commands share, the flag that carries each, and the other spellings every such command also accepts. The rule and its reason are in [ADR 0084](../architecture/decisions/0084-one-concept-one-flag.md). To add a shared flag, add the concept to `libs/actions/flagvocabulary/flag_vocabulary.json` first.\n\n")
	b.WriteString("| Meaning | Flag | Also accepted | Value | Commands |\n|---|---|---|---|---|\n")
	for _, c := range concepts {
		flag := "`--" + c.Flag + "`"
		if c.Short != "" {
			flag = "`-" + c.Short + ", --" + c.Flag + "`"
		}
		var also []string
		for _, s := range c.Synonyms {
			if s.Short != "" {
				also = append(also, "`-"+s.Short+"`")
			}
			also = append(also, "`--"+s.Name+"`")
		}
		commands := "many commands"
		if len(c.Commands) > 0 {
			var list []string
			for _, p := range c.Commands {
				list = append(list, "`smith "+p+"`")
			}
			commands = strings.Join(list, ", ")
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", capitalise(c.Meaning), flag, orDash(strings.Join(also, ", ")), valueWords(c.Type), commands)
	}
	b.WriteString("\nThe synonyms behave exactly like the flag: they can be mixed on one command line, and `--help` shows the flag once.\n")
	if len(exceptions) > 0 {
		b.WriteString("\n## Deliberate exceptions\n\nThese flags look like a shared one but are not.\n\n| Command | Flag | Why |\n|---|---|---|\n")
		for _, e := range exceptions {
			fmt.Fprintf(&b, "| `smith %s` | `--%s` | %s |\n", e.Command, e.Flag, e.Why)
		}
	}
	return b.String()
}

// keyWidth is the width of the longest key plus its colon, so the values of the
// generated object line up as the repository's ESLint key-spacing rule asks.
func keyWidth(concepts []flagvocabulary.Concept) int {
	width := 0
	for _, c := range concepts {
		width = max(width, len(camel(c.ID))+1)
	}
	return width
}

func spelling(c flagvocabulary.Concept) string {
	if c.Short != "" {
		return "-" + c.Short
	}
	return "--" + c.Flag
}

func camel(id string) string {
	parts := strings.Split(id, "-")
	for i := 1; i < len(parts); i++ {
		parts[i] = capitalise(parts[i])
	}
	return strings.Join(parts, "")
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func valueWords(t string) string {
	switch t {
	case "stringSlice":
		return "text, repeatable or comma separated"
	case "bool":
		return "on or off"
	}
	return "text"
}
