package recipe

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// ErrUnknownPlugin is wrapped by the lookup of a plugin that is neither
// built in, declared in smith.toml nor installed on the search path.
var ErrUnknownPlugin = errors.New("no such plugin")

// Catalog finds the plugins a recipe may name: the built-in ones by their
// own description (nothing is started), and the others by starting them for
// their handshake.
type Catalog struct {
	// Root is the workspace root, where a search path and smith.toml are
	// looked for; empty outside a workspace.
	Root string
	// Version is smith's, which a built-in plugin reports as its own.
	Version string
	// External are the plugins smith.toml declares.
	External map[string]workspace.ExternalPlugin
}

// Lookup describes a plugin by name.
func (c Catalog) Lookup(name string) (protocol.InitializeResult, error) {
	_, replaced := c.External[name]
	if ctor, ok := pluginhost.Builtin(name); ok && !replaced {
		p := ctor(c.Version)
		info, err := p.Describe(protocol.InitializeParams{})
		if err != nil {
			return info, err
		}
		info.Capabilities = serve.CapabilitiesOf(p)
		return info, nil
	}
	command, err := pluginhost.LaunchCommand(c.Root, c.External, name)
	if err != nil {
		return protocol.InitializeResult{}, fmt.Errorf("%w: %q is not built in, declared in smith.toml or installed (built in: %s)", ErrUnknownPlugin, name, strings.Join(c.builtinNames(), ", "))
	}
	client, err := host.Start(host.Options{
		Command:       command,
		Env:           append(os.Environ(), "SMITH_PLUGIN=1"),
		WorkspaceRoot: c.Root,
		Stderr:        io.Discard,
		Host:          protocol.Peer{Name: "smith", Version: c.Version},
	})
	if err != nil {
		return protocol.InitializeResult{}, fmt.Errorf("starting plugin %s: %w", name, err)
	}
	defer func() { _ = client.Close() }()
	return client.Info, nil
}

func (c Catalog) builtinNames() []string {
	names := append(pluginhost.BuiltinNames(), sortedKeys(pluginhost.BuiltinAddons)...)
	sort.Strings(names)
	return names
}
