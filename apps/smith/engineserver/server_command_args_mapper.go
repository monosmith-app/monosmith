package engineserver

import (
	"encoding/json"
	"fmt"
	"strings"
)

// settingParam checks the name of a setting of smith.toml ("ci.devcontainer"):
// one word that is not an option.
func settingParam(setting string) *rpcError {
	if setting == "" || strings.HasPrefix(setting, "-") || strings.ContainsAny(setting, " \t\r\n") {
		return &rpcError{codeInvalidParams, `"setting" must be a setting such as "ci.devcontainer"`}
	}
	return nil
}

// releaseArgs maps release's params to `smith release --yes`: the person was
// asked by the client, so the command does not ask again.
func releaseArgs(raw json.RawMessage) ([]string, *rpcError) {
	var p ReleaseParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	args := []string{"release", "--yes"}
	if p.NoPush {
		args = append(args, "--no-push")
	}
	return args, nil
}

// configSetArgs maps config.set's params to `smith config set`. The values
// follow `--`, so one that starts with a dash (a negative number) is a value.
func configSetArgs(raw json.RawMessage) ([]string, *rpcError) {
	var p ConfigSetParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	if rerr := settingParam(p.Setting); rerr != nil {
		return nil, rerr
	}
	values := p.Values
	if len(values) == 0 && p.Value != "" {
		values = []string{p.Value}
	}
	if len(values) == 0 {
		return nil, &rpcError{codeInvalidParams, fmt.Sprintf("config.set needs the value of %s: {\"setting\": %q, \"value\": \"...\"}, or \"values\" for a list", p.Setting, p.Setting)}
	}
	return append([]string{"config", "set", "--", p.Setting}, values...), nil
}

// configUnsetArgs maps config.unset's params to `smith config unset`.
func configUnsetArgs(raw json.RawMessage) ([]string, *rpcError) {
	var p SettingParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	if rerr := settingParam(p.Setting); rerr != nil {
		return nil, rerr
	}
	return []string{"config", "unset", p.Setting}, nil
}
