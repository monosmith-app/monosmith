package manifestoffer

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Written is what Write wrote.
type Written struct {
	// Files are the files written, from the workspace root.
	Files []string
	// Stack is the stack recorded in the manifest, "" for the language plugin alone.
	Stack string
	// Plugin is the plugin that wrote the files.
	Plugin string
}

// Write has a plugin write the manifest of a folder that holds sources: the
// scaffold of the workspace's default stack for the language
// ([default_stacks]), else the language plugin, generates with ManifestOnly,
// so the files come from the same templates as smith add. The sources are
// not touched, and a file that exists is never written over.
//
// Parameters: s is the open session at the workspace root; language is the
// plugin's language; kind, name and dir are the project's kind, name and
// folder (from the workspace root, "." for the root itself).
// Returns: what was written, or a problem when no plugin can write it.
func Write(s *session.Session, language string, kind protocol.ProjectKind, name, dir string) (Written, error) {
	lang, err := s.Plugins.ForLanguage(language)
	if err != nil {
		return Written{}, err
	}
	generator, stack := lang, ""
	if st := s.Config.DefaultStacks[lang.Info.Plugin.Name]; st != "" && st != protocol.NoStack {
		found, err := s.Stacks.Get(st)
		if err != nil {
			return Written{}, err
		}
		stack = found.Name
		if found.Scaffold != "" {
			if g := s.Plugins.Get(found.Scaffold); g != nil {
				generator = g
			}
		}
	}
	if !generator.Info.Capabilities.ManifestOnly {
		return Written{}, problem.New(fmt.Sprintf("the %s plugin cannot write only a manifest into a folder that holds sources", generator.Info.Plugin.Name),
			"write the manifest yourself, then run the command again")
	}
	params := protocol.GenerateParams{Kind: kind, Name: name, Directory: dir, Stack: stack, ManifestOnly: true, Workspace: s.Config.Workspace.Info(s.Config.Workspace.ModulePrefix)}
	var res protocol.GenerateResult
	if err := generator.Call(protocol.MethodGenerate, params, &res, 5*time.Minute); err != nil {
		return Written{}, fmt.Errorf("the %s plugin did not write the manifest: %w", generator.Info.Plugin.Name, err)
	}
	return Written{Files: res.Files, Stack: stack, Plugin: generator.Info.Plugin.Name}, nil
}
