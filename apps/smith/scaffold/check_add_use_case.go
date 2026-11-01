package scaffold

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckAdd refuses an add that cannot succeed, before the plan is shown: a
// kind that does not exist, a bad or taken name, choices that do not fit
// together, a like-project that is not there, and a generator that cannot make
// the kind. It asks nothing and writes nothing, and it skips the checks that
// need an answer a person has still to give (a recipe, a language to pick).
//
// Parameters: s is the open session; opts are the options smith add runs with.
// Returns: the first problem found, or nil.
func CheckAdd(s *session.Session, opts AddOptions) error {
	kind := protocol.ProjectKind(opts.Kind)
	switch kind {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
	default:
		return fmt.Errorf("kind %q is not app, lib or package. To add a package (a dependency) to a project, use: smith install -w <project> <package>", opts.Kind)
	}
	if err := ValidateName(opts.Kind, opts.Name); err != nil {
		return err
	}
	if p := s.Graph.Get(opts.Name); p != nil {
		return fmt.Errorf("a project named %q already exists (%s)", opts.Name, p.Root)
	}
	if _, err := likeProject(s, opts); err != nil {
		return err
	}
	var language *host.Client
	var st protocol.Stack
	switch {
	case len(opts.With) > 0:
		plan, err := planSelection(s, opts)
		if err != nil {
			return err
		}
		language, st = s.Plugins.Get(plan.stack.Language), plan.stack
	case (opts.Stack != "" || opts.Lang != "") && (opts.Stack != "" || s.Config.Workspace.Recipe == ""):
		var err error
		if language, st, err = chooseStack(s, opts); err != nil {
			return err
		}
	default:
		return nil
	}
	if language == nil {
		return nil
	}
	generator := language
	if st.Scaffold != "" {
		if g := s.Plugins.Get(st.Scaffold); g != nil {
			generator = g
		}
	}
	return checkScaffolds(generator, kind)
}

// checkScaffolds refuses a generator that cannot make a project of the kind.
func checkScaffolds(generator *host.Client, kind protocol.ProjectKind) error {
	if !generator.Supports(protocol.MethodGenerate) {
		return fmt.Errorf("plugin %s cannot generate projects", generator.Info.Plugin.Name)
	}
	if generator.Info.Scaffolds(kind) {
		return nil
	}
	var can []string
	for _, k := range generator.Info.ScaffoldKinds() {
		can = append(can, string(k))
	}
	return fmt.Errorf("%s cannot scaffold a %s; it scaffolds %s", generator.Info.Plugin.Name, kind, strings.Join(can, " and "))
}
