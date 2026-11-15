package nxworkspace

import (
	"encoding/json"
	"fmt"
)

// Target is one Nx target.
type Target struct {
	Executor string `json:"executor"`
	// Command is the shorthand for nx:run-commands with one command.
	Command        string                     `json:"command"`
	Options        map[string]json.RawMessage `json:"options"`
	Configurations map[string]json.RawMessage `json:"configurations"`
	DependsOn      []DependsOn                `json:"dependsOn"`
	Inputs         []Input                    `json:"inputs"`
	Outputs        []string                   `json:"outputs"`
	Cache          *bool                      `json:"cache"`
	Continuous     bool                       `json:"continuous"`
	Parallelism    *bool                      `json:"parallelism"`
	// Metadata.Description is Nx's description of an inferred target.
	Metadata struct {
		Description string `json:"description"`
	} `json:"metadata"`
}

// overlay returns t with the fields other sets on top.
func (t Target) overlay(other Target) Target {
	if other.Executor != "" {
		t.Executor = other.Executor
	}
	if other.Command != "" {
		t.Command = other.Command
	}
	if other.Options != nil {
		t.Options = other.Options
	}
	if other.Configurations != nil {
		t.Configurations = other.Configurations
	}
	if other.DependsOn != nil {
		t.DependsOn = other.DependsOn
	}
	if other.Inputs != nil {
		t.Inputs = other.Inputs
	}
	if other.Outputs != nil {
		t.Outputs = other.Outputs
	}
	if other.Cache != nil {
		t.Cache = other.Cache
	}
	if other.Continuous {
		t.Continuous = true
	}
	if other.Parallelism != nil {
		t.Parallelism = other.Parallelism
	}
	return t
}

// Option decodes one option into out; it reports whether the option exists.
func (t Target) Option(name string, out any) bool {
	raw, ok := t.Options[name]
	if !ok {
		return false
	}
	return json.Unmarshal(raw, out) == nil
}

// DependsOn is "target", "^target", or {target, projects, dependencies}.
type DependsOn struct {
	Target string
	// Dependencies means the target of every dependency ("^target").
	Dependencies bool
	// Projects names the projects whose Target runs first: Nx names,
	// "self", "dependencies", or patterns.
	Projects []string
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DependsOn) UnmarshalJSON(raw []byte) error {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		if len(s) > 0 && s[0] == '^' {
			d.Target, d.Dependencies = s[1:], true
		} else {
			d.Target = s
		}
		return nil
	}
	var obj struct {
		Target       string          `json:"target"`
		Dependencies bool            `json:"dependencies"`
		Projects     json.RawMessage `json:"projects"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("dependsOn: %w", err)
	}
	d.Target, d.Dependencies = obj.Target, obj.Dependencies
	if len(obj.Projects) > 0 {
		var list StringList
		if json.Unmarshal(obj.Projects, &list) == nil {
			d.Projects = list
		}
	}
	return nil
}

// Input is a file glob, a named input, or an object ({env}, {runtime},
// {externalDependencies}, {fileset}, {input, projects|dependencies}).
type Input struct {
	// Text is the string form: a glob with {projectRoot}/{workspaceRoot}
	// tokens, a named input, or "^named" for dependencies' inputs.
	Text string
	// Env, Runtime and Fileset are the object forms.
	Env     string
	Runtime string
	Fileset string
	// Named is {input: "name"}; Dependencies marks it as "^name".
	Named        string
	Dependencies bool
	// External lists {externalDependencies: [...]}.
	External []string
}

// UnmarshalJSON implements json.Unmarshaler.
func (in *Input) UnmarshalJSON(raw []byte) error {
	if json.Unmarshal(raw, &in.Text) == nil {
		return nil
	}
	var obj struct {
		Env                  string   `json:"env"`
		Runtime              string   `json:"runtime"`
		Fileset              string   `json:"fileset"`
		Input                string   `json:"input"`
		Dependencies         bool     `json:"dependencies"`
		Projects             any      `json:"projects"`
		ExternalDependencies []string `json:"externalDependencies"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("inputs: %w", err)
	}
	in.Env, in.Runtime, in.Fileset, in.External = obj.Env, obj.Runtime, obj.Fileset, obj.ExternalDependencies
	in.Named, in.Dependencies = obj.Input, obj.Dependencies
	return nil
}
