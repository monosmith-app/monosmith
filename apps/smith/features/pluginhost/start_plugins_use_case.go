package pluginhost

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/features/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginenv"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Options configure Start.
type Options struct {
	WorkspaceRoot string
	Config        workspace.Config
	// Version is the host version, also handed to built-in plugins.
	Version string
	// Log receives plugin stderr.
	Log io.Writer
}

// Start launches every enabled plugin in parallel and initializes it. A
// built-in plugin is served inside this process; an external one runs as its
// own process with the minimal environment built below.
func Start(opts Options) (*PluginSet, error) {
	log := opts.Log
	if log == nil {
		log = os.Stderr
	}
	// A plugin's stdout is the protocol pipe, so it cannot tell whether the
	// tools it runs reach a person who sees colour: the host decides for it.
	colour := termcolour.ToolEntry(termcolour.ToolWantedFor(os.Getenv, log, term.IsTerminal))
	// Built-in plugins run inside this process (one process per session, not
	// one per plugin): the tools they start read the colour decision from the
	// environment, and their output goes to the session's log.
	if key, value, ok := strings.Cut(colour, "="); ok && os.Getenv(key) == "" {
		_ = os.Setenv(key, value)
	}
	toolrun.Log = log
	type launch struct {
		name    string
		command []string
		env     []string
		// builtin serves the plugin in this process; nil for an external one.
		builtin func(version string) serve.Plugin
	}
	var launches []launch
	enabled := opts.Config.Plugins.Enabled
	if len(enabled) == 0 {
		enabled = BuiltinNames()
	}
	for _, name := range enabled {
		if _, ok := Builtins[name]; !ok {
			return nil, fmt.Errorf("smith.toml enables unknown plugin %q; built-in plugins are %v", name, BuiltinNames())
		}
		if _, replaced := opts.Config.Plugins.External[name]; replaced {
			continue
		}
		launches = append(launches, launch{name: name, builtin: Builtins[name]})
	}
	var approved []pluginadmit.Plugin
	external := make([]string, 0, len(opts.Config.Plugins.External))
	for name := range opts.Config.Plugins.External {
		external = append(external, name)
	}
	sort.Strings(external)
	for _, name := range external {
		p, err := ResolveExternal(opts.WorkspaceRoot, opts.Config.Plugins.External, name)
		if err != nil {
			return nil, fmt.Errorf("smith.toml [plugins.external.%s]: %w", name, err)
		}
		approved = append(approved, p)
		env := append(pluginenv.Build(os.Environ(), p.ManifestEnv, p.WorkspaceEnv, runtime.GOOS), "SMITH_PLUGIN=1", colour)
		launches = append(launches, launch{name: name, command: p.Command, env: env})
	}
	// No plugin starts until every external one the workspace needs is
	// trusted.
	if err := pluginadmit.Admit(opts.WorkspaceRoot, approved); err != nil {
		return nil, err
	}
	for _, name := range sortedNames(opts.Config.Conventions) {
		_, external := opts.Config.Plugins.External[name]
		_, builtin := BuiltinAddons[name]
		if !external && !builtin {
			return nil, fmt.Errorf("smith.toml configures unknown convention %q; built-in ones are %v, and an external one is declared under [plugins.external.%s]", name, sortedNames(BuiltinAddons), name)
		}
	}
	for _, name := range sortedNames(BuiltinAddons) {
		if _, ok := opts.Config.Plugins.External[name]; ok {
			continue
		}
		launches = append(launches, launch{name: name, builtin: BuiltinAddons[name]})
	}

	clients := make([]*host.Client, len(launches))
	errs := make([]error, len(launches))
	var wg sync.WaitGroup
	for i, l := range launches {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hostOpts := host.Options{
				Command:       l.command,
				Env:           l.env,
				WorkspaceRoot: opts.WorkspaceRoot,
				Stderr:        log,
				Host:          protocol.Peer{Name: "smith", Version: opts.Version},
			}
			if l.builtin != nil {
				clients[i], errs[i] = startInProcess(l.builtin(opts.Version), hostOpts)
				return
			}
			clients[i], errs[i] = host.Start(hostOpts)
		}()
	}
	wg.Wait()
	set := &PluginSet{}
	for i, c := range clients {
		if errs[i] != nil {
			set.Clients = append(set.Clients, clients[:i]...)
			set.Close()
			return nil, fmt.Errorf("plugin %s: %w", launches[i].name, errs[i])
		}
		set.Clients = append(set.Clients, c)
	}
	for _, name := range external {
		if err := checkExternal(set, name, opts.Config.Plugins.External[name]); err != nil {
			set.Close()
			return nil, err
		}
	}
	return set, nil
}

// startInProcess serves a built-in plugin in this process over two pipes, with
// the same protocol a plugin process speaks: closing the client ends it.
func startInProcess(p serve.Plugin, opts host.Options) (*host.Client, error) {
	fromHost, toPlugin := io.Pipe()
	fromPlugin, toHost := io.Pipe()
	go func() {
		err := serve.Serve(p, fromHost, toHost)
		_ = toHost.CloseWithError(err)
		_ = fromHost.CloseWithError(err)
	}()
	return host.Connect(fromPlugin, toPlugin, opts)
}

// checkExternal checks that an external plugin answered with the name it is
// declared under and a version smith.toml accepts.
func checkExternal(set *PluginSet, name string, cfg workspace.ExternalPlugin) error {
	c := set.Get(name)
	if c == nil {
		return fmt.Errorf("smith.toml [plugins.external.%s]: the plugin calls itself something else; declare it under its own name", name)
	}
	ok, err := Satisfies(c.Info.Plugin.Version, cfg.Version)
	if err != nil {
		return fmt.Errorf("smith.toml [plugins.external.%s]: %w", name, err)
	}
	if !ok {
		return fmt.Errorf("smith.toml [plugins.external.%s] wants version %s, but the plugin is %s", name, cfg.Version, c.Info.Plugin.Version)
	}
	return nil
}
