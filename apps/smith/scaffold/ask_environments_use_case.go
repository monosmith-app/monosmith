package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// NoEnvironments is the value of --environments that asks for a single build.
const NoEnvironments = "none"

// EnvironmentFlags are the flags of smith add that answer the questions about
// the environment builds of a web app (ADR 0081). A field that is empty was not
// given: it is asked, or takes its default.
type EnvironmentFlags struct {
	// Environments is --environments: the names, or NoEnvironments.
	Environments []string
	// Default is --default-environment.
	Default string
	// Output is --environment-output: workspace or project.
	Output string
	// EnvFiles is --env-files: commit or example.
	EnvFiles string
}

// EnvironmentsFlagHelp names the flags in the message that stops an add without
// a terminal, so the person sees what to type.
const EnvironmentsFlagHelp = "--environments lcl,dev,uat,prd (or --environments none for one build), and when you want another answer: --default-environment <name>, --environment-output workspace|project, --env-files commit|example"

// scaffoldPlugin is the plugin that writes the project an add will make: the
// scaffold of its stack, or of its choices. It returns "" when the stack is not
// known before the add (a recipe asks its own questions) or cannot be found.
func scaffoldPlugin(s *session.Session, opts AddOptions) string {
	if len(opts.With) > 0 {
		plan, err := planSelection(s, opts)
		if err != nil || plan == nil {
			return ""
		}
		return plan.stack.Scaffold
	}
	if opts.Recipe != nil && len(opts.Recipe.Answers) > 0 || opts.Stack == "" && s.Config.Workspace.Recipe != "" {
		return ""
	}
	_, st, err := chooseStack(s, opts)
	if err != nil {
		return ""
	}
	return st.Scaffold
}

// CompleteEnvironments settles the environment builds of an app that is about
// to be added. A web app (a Vite scaffold) is built once per environment
// because Vite writes its VITE_ variables into the bundle when it builds; the
// names, the default environment, where the output goes and whether the env
// files are committed are each a choice, with the default marked.
//
// With a terminal and no --environments it asks, pre-filling the names from
// [workspace] environments (else lcl, dev, uat, prd) so the person edits them.
// Without a terminal, or with --yes, --environments is required: it stops and
// names the flags. smith never builds a web app without saying which
// environments it has.
//
// Parameters: s the open session; opts the options of the add; f the flags;
// in and out the terminal; interactive says a person is there to answer.
// Returns: the settings, or nil when the project is not a web app or the
// person asked for a single build; an error for names that are not valid, for
// --environments on a project that is not a web app, and for a missing
// --environments where nobody can be asked.
func CompleteEnvironments(s *session.Session, opts AddOptions, f EnvironmentFlags, in io.Reader, out io.Writer, interactive bool) (*webenv.Settings, error) {
	given := len(f.Environments) > 0
	if protocolKindIsNotApp(opts.Kind) || scaffoldPlugin(s, opts) != webenv.PluginName {
		if given || f.Default != "" || f.Output != "" || f.EnvFiles != "" {
			return nil, fmt.Errorf("the environment flags are for a web app built with Vite (smith add app web --with react); %s is not one", opts.Name)
		}
		return nil, nil
	}
	if given && len(f.Environments) == 1 && f.Environments[0] == NoEnvironments {
		return nil, nil
	}
	if !given && !interactive {
		return nil, fmt.Errorf("%s is a web app, and a web app is built once per environment: Vite writes its VITE_ values into the bundle when it builds. Say which environments it has: %s", opts.Name, EnvironmentsFlagHelp)
	}
	reader := bufio.NewReader(in)
	names := f.Environments
	defaults := s.Config.Workspace.Environments
	if len(defaults) == 0 {
		defaults = webenv.DefaultNames
	}
	if !given {
		fmt.Fprintf(out, "\n%s is a web app. Vite writes its VITE_ values into the bundle when it builds, so smith builds it once per environment, each with its own .env.<name> file.\n", opts.Name)
		fmt.Fprintf(out, "Environments (lowercase words, separated by commas) [%s]: ", strings.Join(defaults, ","))
		line, _ := reader.ReadString('\n')
		if typed := strings.TrimSpace(line); typed != "" {
			names = webenv.ParseList(typed)
		} else {
			names = slices.Clone(defaults)
		}
	}
	settings := webenv.Settings{Names: names, Default: f.Default, Output: f.Output, EnvFiles: f.EnvFiles}
	if err := webenv.ValidateNames(names); err != nil {
		return nil, err
	}
	if interactive && !given {
		settings = askEnvironmentChoices(settings, reader, out)
	}
	settings = settings.Normalized()
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	return &settings, nil
}

// protocolKindIsNotApp reports whether the kind is not an app: only an app is a
// web app.
func protocolKindIsNotApp(kind string) bool { return kind != "app" }

// askEnvironmentChoices asks what the flags left open: the default environment,
// where the builds go and whether the env files are committed. Each answer has
// its default marked and its consequence said.
func askEnvironmentChoices(s webenv.Settings, in *bufio.Reader, out io.Writer) webenv.Settings {
	if s.Default == "" {
		fallback := s.Normalized().Default
		var options []choiceprompt.Option
		for _, n := range s.Names {
			hint := ""
			if n == fallback {
				hint = "(default)"
			}
			options = append(options, choiceprompt.Option{Label: n, Value: n, Hint: hint})
		}
		s.Default = choiceprompt.Ask("Which environment is `smith run build` (what CI and the release build)? The others are build:<name>.", options, fallback, in, out)
	}
	if s.Output == "" {
		s.Output = choiceprompt.Ask("Where do the builds go?", []choiceprompt.Option{
			{Label: "workspace", Value: webenv.OutputWorkspace, Hint: "dist/apps/<app>/<name> at the workspace root, with every other build (default)"},
			{Label: "project", Value: webenv.OutputProject, Hint: "dist/<name> inside the app's own folder; smith doctor then accepts it"},
		}, webenv.OutputWorkspace, in, out)
	}
	if s.EnvFiles == "" {
		s.EnvFiles = choiceprompt.Ask("The .env.<name> files hold VITE_ values, which end up in the browser, so they hold no secret. Commit them?", []choiceprompt.Option{
			{Label: "commit", Value: webenv.FilesCommit, Hint: "git keeps .env.<name>; CI builds with them (default)"},
			{Label: "example", Value: webenv.FilesExample, Hint: "git keeps only .env.example; each machine and CI job makes its own files"},
		}, webenv.FilesCommit, in, out)
	}
	return s
}
