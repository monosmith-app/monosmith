package turboworkspace

import (
	"encoding/json"
	"sort"
)

// TurboJSON is the part of turbo.json a conversion reads. Tasks holds `tasks`
// (Turborepo 2) or, failing that, the older `pipeline`.
type TurboJSON struct {
	// File is the workspace-relative path it was read from.
	File    string   `json:"-"`
	Extends []string `json:"extends"`
	// Tasks are the task definitions, keyed by "task", "package#task" or
	// "//#task".
	Tasks                map[string]TaskConfig `json:"tasks"`
	Pipeline             map[string]TaskConfig `json:"pipeline"`
	GlobalDependencies   []string              `json:"globalDependencies"`
	GlobalEnv            []string              `json:"globalEnv"`
	GlobalPassThroughEnv []string              `json:"globalPassThroughEnv"`
	RemoteCache          json.RawMessage       `json:"remoteCache"`
	// Keys are the top-level keys the file has, so an option smith does not
	// know is reported rather than dropped silently.
	Keys []string `json:"-"`
}

// TaskConfig is one turbo.json task definition. A nil slice or pointer means
// the field is not set, which a package configuration needs to know: only a
// field it sets overrides the root's.
type TaskConfig struct {
	DependsOn      []string `json:"dependsOn"`
	Outputs        []string `json:"outputs"`
	Inputs         []string `json:"inputs"`
	Env            []string `json:"env"`
	PassThroughEnv []string `json:"passThroughEnv"`
	Cache          *bool    `json:"cache"`
	Persistent     *bool    `json:"persistent"`
	Interactive    *bool    `json:"interactive"`
	// OutputLogs is `outputLogs`, or `outputMode`, its Turborepo 1 name.
	OutputLogs string   `json:"outputLogs"`
	OutputMode string   `json:"outputMode"`
	With       []string `json:"with"`
	Extends    *bool    `json:"extends"`
	// Keys are the fields the definition has, to report unknown ones.
	Keys []string `json:"-"`
}

// UnmarshalJSON records the fields present, besides decoding them.
func (t *TaskConfig) UnmarshalJSON(raw []byte) error {
	type plain TaskConfig
	if err := json.Unmarshal(raw, (*plain)(t)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for k := range fields {
		t.Keys = append(t.Keys, k)
	}
	sort.Strings(t.Keys)
	return nil
}

// TaskSet returns the task definitions: `tasks`, else `pipeline`.
func (t TurboJSON) TaskSet() map[string]TaskConfig {
	if len(t.Tasks) > 0 {
		return t.Tasks
	}
	return t.Pipeline
}
