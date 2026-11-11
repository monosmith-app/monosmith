package depsrepair

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// installTimeout bounds one install: a cold cache downloads every package.
const installTimeout = 20 * time.Minute

// SessionPorts are the ports of a real repair in an open workspace: the
// link-safe DeleteTree, the commands run as the doctor's fix steps run (the
// same program lookup, so npm.cmd is found on Windows), and the install of the
// plugin of each tool.
//
// Parameters: s is the open workspace; it must stay open while the ports are used.
//
// Returns: the ports.
func SessionPorts(s *session.Session) Ports {
	return Ports{
		Delete: DeleteTree,
		Run: func(ctx context.Context, root string, step protocol.FixStep, out io.Writer) error {
			return doctorfix.RunSteps(ctx, root, []protocol.FixStep{step}, out)
		},
		Install: func(ctx context.Context, tool Tool, out io.Writer) error {
			for _, c := range s.Plugins.Clients {
				if c.Info.Plugin.Name != tool.Plugin {
					continue
				}
				var res protocol.DependencyResult
				if err := c.CallContext(ctx, protocol.MethodInstall, protocol.InstallParams{Projects: tool.Projects}, &res, installTimeout); err != nil {
					return err
				}
				for _, n := range res.Notes {
					fmt.Fprintln(out, "  "+n)
				}
				return nil
			}
			return fmt.Errorf("the plugin %s is not loaded", tool.Plugin)
		},
	}
}
