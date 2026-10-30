package release

import (
	"context"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteConventionVersions records version in the files that the project's
// other plugins (its stack's toolchain and its conventions, such as tauri's
// tauri.conf.json and Cargo.toml) keep a version in, beside the manifest its
// language plugin writes. It returns the files written, or that would be in a
// dry run.
func WriteConventionVersions(ctx context.Context, s *session.Session, p *projectgraph.Project, version string, dryRun bool) ([]string, error) {
	var files []string
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		if name == p.Plugin || !p.ServedBy(name) || !c.Supports(protocol.MethodVersion) {
			continue
		}
		var vr protocol.VersionResult
		if err := c.CallContext(ctx, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: version, DryRun: dryRun}, &vr, time.Minute); err != nil {
			return files, err
		}
		files = append(files, vr.Files...)
	}
	return files, nil
}
