package nxworkspace

import (
	"encoding/json"
	"fmt"
)

// NxJSON is the part of nx.json a conversion reads.
type NxJSON struct {
	NamedInputs    map[string][]Input `json:"namedInputs"`
	TargetDefaults TargetDefaults     `json:"targetDefaults"`
	Plugins        []PluginEntry      `json:"plugins"`
	Release        *Release           `json:"release"`
	DefaultBase    string             `json:"defaultBase"`
	NxCloudID      string             `json:"nxCloudId"`
	NxCloudToken   string             `json:"nxCloudAccessToken"`
	// Mnci is the settings block the old mnci (4.x) keeps in nx.json.
	Mnci *LegacySettings `json:"mnci"`
}

// TargetDefaults maps a target name to its defaults. Nx 24 also allows a list
// of entries for one target (each may be filtered by project or plugin); the
// entries are read and merged in order, a filter is not applied, because the
// project graph smith reads from Nx has the defaults applied already.
type TargetDefaults map[string]Target

// UnmarshalJSON accepts {"test": {...}} and {"test": [{...}, {...}]}.
func (d *TargetDefaults) UnmarshalJSON(raw []byte) error {
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("targetDefaults should be an object keyed by target name: %w", err)
	}
	out := make(TargetDefaults, len(entries))
	for name, entry := range entries {
		var list []Target
		if json.Unmarshal(entry, &list) == nil {
			var merged Target
			for _, t := range list {
				merged = merged.overlay(t)
			}
			out[name] = merged
			continue
		}
		var one Target
		if err := json.Unmarshal(entry, &one); err != nil {
			return fmt.Errorf("targetDefaults.%s should be an object or a list of objects: %w", name, err)
		}
		out[name] = one
	}
	*d = out
	return nil
}

// PluginEntry is a plugin name, with options when nx.json gives an object.
type PluginEntry struct {
	Plugin  string                     `json:"plugin"`
	Options map[string]json.RawMessage `json:"options"`
	Include []string                   `json:"include"`
	Exclude []string                   `json:"exclude"`
}

// UnmarshalJSON accepts "name" and {"plugin": "name", ...}.
func (p *PluginEntry) UnmarshalJSON(raw []byte) error {
	var name string
	if json.Unmarshal(raw, &name) == nil {
		p.Plugin = name
		return nil
	}
	type plain PluginEntry
	return json.Unmarshal(raw, (*plain)(p))
}

// Release is nx.json's release section.
type Release struct {
	Projects             StringList                 `json:"projects"`
	ProjectsRelationship string                     `json:"projectsRelationship"`
	Groups               map[string]json.RawMessage `json:"groups"`
	ReleaseTag           struct {
		Pattern string `json:"pattern"`
	} `json:"releaseTag"`
	// ReleaseTagPattern is the pre-21 spelling of releaseTag.pattern.
	ReleaseTagPattern string `json:"releaseTagPattern"`
	Version           struct {
		ConventionalCommits bool `json:"conventionalCommits"`
	} `json:"version"`
	Changelog struct {
		WorkspaceChangelog json.RawMessage `json:"workspaceChangelog"`
		ProjectChangelogs  json.RawMessage `json:"projectChangelogs"`
	} `json:"changelog"`
}

// TagPattern is the release tag pattern, Nx's default when unset.
func (r *Release) TagPattern() string {
	if r == nil {
		return "{projectName}@{version}"
	}
	if r.ReleaseTag.Pattern != "" {
		return r.ReleaseTag.Pattern
	}
	if r.ReleaseTagPattern != "" {
		return r.ReleaseTagPattern
	}
	return "{projectName}@{version}"
}

// CreatesGitHubReleases reports whether changelogs create GitHub Releases.
func (r *Release) CreatesGitHubReleases() bool {
	if r == nil {
		return false
	}
	for _, raw := range []json.RawMessage{r.Changelog.ProjectChangelogs, r.Changelog.WorkspaceChangelog} {
		var c struct {
			CreateRelease json.RawMessage `json:"createRelease"`
		}
		if json.Unmarshal(raw, &c) != nil || len(c.CreateRelease) == 0 {
			continue
		}
		var provider string
		if json.Unmarshal(c.CreateRelease, &provider) == nil && provider == "github" {
			return true
		}
		var obj struct {
			Provider string `json:"provider"`
		}
		if json.Unmarshal(c.CreateRelease, &obj) == nil && obj.Provider == "github" {
			return true
		}
	}
	return false
}

// LegacySettings is the old mnci's block in nx.json.
type LegacySettings struct {
	WorkspaceName string `json:"workspaceName"`
	Scope         string `json:"scope"`
	Registry      struct {
		Kind string `json:"kind"`
	} `json:"registry"`
	// CI is one provider, or a list of them.
	CI    StringList `json:"ci"`
	Agent string     `json:"agent"`
	Stack struct {
		TestRunner string `json:"testRunner"`
	} `json:"stack"`
}

// StringList accepts a string or a list of strings.
type StringList []string

// UnmarshalJSON implements json.Unmarshaler.
func (l *StringList) UnmarshalJSON(raw []byte) error {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		*l = StringList{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return fmt.Errorf("want a string or a list of strings: %w", err)
	}
	*l = many
	return nil
}
