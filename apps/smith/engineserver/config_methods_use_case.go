package engineserver

import (
	"context"
	"encoding/json"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	smithworkspace "github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// configList answers with the settings of smith.toml, their values and what
// they do: what `smith config list --json` prints.
func (s *Server) configList(ctx context.Context, _ json.RawMessage) (any, *rpcError) {
	return s.queryArgs(ctx, []string{"config", "list", "--json"})
}

// configGet answers with one setting's value as written in smith.toml, and
// whether it is set at all (when it is not, it has its default).
func (s *Server) configGet(_ context.Context, raw json.RawMessage) (any, *rpcError) {
	var p SettingParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	if rerr := settingParam(p.Setting); rerr != nil {
		return nil, rerr
	}
	root, err := smithworkspace.FindRoot(s.dir())
	if err != nil {
		return nil, &rpcError{codeFailed, err.Error()}
	}
	field, err := configedit.Find(p.Setting)
	if err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	value, set, err := configedit.Get(root, p.Setting)
	if err != nil {
		return nil, &rpcError{codeFailed, err.Error()}
	}
	return ConfigValue{Setting: p.Setting, Value: value, Set: set, Type: field.Type, Description: field.Description}, nil
}
