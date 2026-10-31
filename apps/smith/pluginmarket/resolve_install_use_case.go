package pluginmarket

import (
	"context"
	"net/http"

	"github.com/monosmith-app/monosmith/apps/smith/plugininstall"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// NewResolver is the plugininstall.Resolver for the marketplace at location
// (see Location). It reads the index only when the argument names a plugin
// (see ParseName), so installing from a path or a URL never touches the
// network for this. The archive it returns still goes through the whole
// install check, with the index's digest as the one the archive must match.
// Nothing is trusted: that stays the person's decision.
func NewResolver(client *http.Client, location string, token func() string) plugininstall.Resolver {
	return func(ctx context.Context, arg string) (plugininstall.Resolved, bool, error) {
		name, version, ok := ParseName(arg)
		if !ok {
			return plugininstall.Resolved{}, false, nil
		}
		idx, err := FetchIndex(ctx, client, location, token)
		if err != nil {
			return plugininstall.Resolved{}, false, err
		}
		_, rel, err := Pick(idx, name, version)
		if err != nil {
			return plugininstall.Resolved{}, false, problem.New(err.Error(),
				"`smith plugin search` lists the plugins",
				"a path, URL or github:owner/repo@tag installs without the marketplace")
		}
		return plugininstall.Resolved{
			Name:   name + "@" + rel.Version,
			Origin: location,
			Source: rel.URL,
			SHA256: rel.SHA256,
		}, true, nil
	}
}
