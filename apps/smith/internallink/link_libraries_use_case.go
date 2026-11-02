package internallink

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const timeout = 20 * time.Minute

// LinkOptions control one run of Link.
type LinkOptions struct {
	// Spec is the version range the user passed (--range); empty lets the
	// plugin follow the workspace, or the user choose.
	Spec string
	// Ask puts a real choice to the user; nil when there is no terminal.
	Ask Asker
	// Guard runs change and puts the manifest and lock files in dirs back if it
	// fails, as a failed install does.
	Guard func(dirs []string, change func() error) error
	// Log receives the plan and the result.
	Log io.Writer
}

// ManualSentence is the one sentence for a language whose plugin cannot link:
// it names the plugin and the manifest, and says to add the dependency by hand.
func ManualSentence(plugin string, manifests []string, from *projectgraph.Project) string {
	file := from.Root + "/" + strings.Join(manifests, " or ")
	if len(manifests) == 0 {
		file = from.Root
	}
	return fmt.Sprintf("the %s plugin does not link an internal library yet: add the dependency in %s by hand", plugin, file)
}

// Link makes from depend on each library, after refusing a link that cannot
// work. For each one it asks the plugin for the plan, lets the user choose a
// version range where there is a real choice, shows what will be written to
// which file, and then has the plugin write it. A failed link puts the
// manifest and lock files back.
func Link(s *session.Session, from *projectgraph.Project, libs []*projectgraph.Project, o LinkOptions) error {
	c := s.Plugins.Get(from.Plugin)
	if c == nil || !c.Supports(protocol.MethodLink) {
		var manifests []string
		if c != nil {
			manifests = c.Info.Manifests
		}
		return fmt.Errorf("%s", ManualSentence(from.Plugin, manifests, from))
	}
	for _, lib := range libs {
		if err := Refuse(s.Graph, from, lib); err != nil {
			return err
		}
	}
	for _, lib := range libs {
		params := protocol.LinkParams{Project: from.Project, Library: lib.Project, Plan: true, Spec: o.Spec}
		var plan protocol.LinkResult
		if err := c.Call(protocol.MethodLink, params, &plan, timeout); err != nil {
			return fmt.Errorf("%s (plugin %s): %w", from.Name, from.Plugin, err)
		}
		if len(plan.Options) > 1 && o.Spec == "" {
			spec, err := chooseSpec(from, lib, plan.Options, o)
			if err != nil {
				return err
			}
			params.Spec = spec
			plan = protocol.LinkResult{}
			if err := c.Call(protocol.MethodLink, params, &plan, timeout); err != nil {
				return fmt.Errorf("%s (plugin %s): %w", from.Name, from.Plugin, err)
			}
		}
		fmt.Fprintf(o.Log, "%s will use %s: smith adds %s to %s.\n", from.Name, lib.Name, plan.Line, plan.Manifest)
		params.Plan, params.Quiet = false, true
		var res protocol.LinkResult
		dirs := []string{filepath.Join(s.Root, filepath.FromSlash(from.Root)), s.Root}
		err := o.Guard(dirs, func() error {
			if err := c.Call(protocol.MethodLink, params, &res, timeout); err != nil {
				return fmt.Errorf("%s (plugin %s): %w", from.Name, from.Plugin, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		line := "✔ " + from.Name + ": linked " + lib.Name
		if len(res.Files) > 0 {
			line += " (" + joinNames(res.Files) + ")"
		}
		fmt.Fprintln(o.Log, line)
		for _, n := range res.Notes {
			fmt.Fprintln(o.Log, "  "+n)
		}
	}
	return nil
}

// chooseSpec puts the plugin's range options to the user, or takes the
// default (the first) and says so when there is no terminal.
func chooseSpec(from, lib *projectgraph.Project, options []protocol.LinkOption, o LinkOptions) (string, error) {
	if o.Ask == nil {
		fmt.Fprintf(o.Log, "%s: no terminal to ask in, so smith writes %s (%s). To choose another: --range <range>\n", from.Name, options[0].Spec, options[0].Label)
		return options[0].Spec, nil
	}
	shown := make([]Option, len(options))
	for i, opt := range options {
		shown[i] = Option{Label: opt.Label, Meaning: opt.Meaning}
	}
	pick, err := o.Ask.Choose(fmt.Sprintf("Which version of %s should %s ask for? The workspace has no habit for this yet.", lib.Name, from.Name), shown)
	if err != nil {
		return "", err
	}
	return options[pick].Spec, nil
}

// Unlink removes the dependency of from on each library and tells when there
// was nothing to remove. It returns the libraries that were not linked, so the
// caller may treat their names as registry packages.
func Unlink(s *session.Session, from *projectgraph.Project, libs []*projectgraph.Project, o LinkOptions) ([]*projectgraph.Project, error) {
	c := s.Plugins.Get(from.Plugin)
	if c == nil || !c.Supports(protocol.MethodUnlink) {
		var manifests []string
		if c != nil {
			manifests = c.Info.Manifests
		}
		return nil, fmt.Errorf("%s", strings.Replace(ManualSentence(from.Plugin, manifests, from), "add the dependency", "remove the dependency", 1))
	}
	var notLinked []*projectgraph.Project
	for _, lib := range libs {
		var res protocol.DependencyResult
		dirs := []string{filepath.Join(s.Root, filepath.FromSlash(from.Root)), s.Root}
		err := o.Guard(dirs, func() error {
			if err := c.Call(protocol.MethodUnlink, protocol.UnlinkParams{Project: from.Project, Library: lib.Project}, &res, timeout); err != nil {
				return fmt.Errorf("%s (plugin %s): %w", from.Name, from.Plugin, err)
			}
			return nil
		})
		if err != nil {
			return notLinked, err
		}
		if len(res.Files) == 0 {
			fmt.Fprintf(o.Log, "%s did not depend on %s, so nothing was removed.\n", from.Name, lib.Name)
			notLinked = append(notLinked, lib)
			continue
		}
		fmt.Fprintf(o.Log, "✔ %s: unlinked %s (%s)\n", from.Name, lib.Name, joinNames(res.Files))
		for _, n := range res.Notes {
			fmt.Fprintln(o.Log, "  "+n)
		}
	}
	return notLinked, nil
}
