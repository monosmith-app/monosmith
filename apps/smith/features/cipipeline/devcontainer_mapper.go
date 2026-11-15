package cipipeline

import (
	"encoding/json"
	"sort"
	"strings"
)

// DevcontainerPath is where the dev container definition is written.
const DevcontainerPath = ".devcontainer/devcontainer.json"

// DevcontainerImage is the base image; every toolchain comes from a feature
// at the version CI installs.
const DevcontainerImage = "mcr.microsoft.com/devcontainers/base:bookworm"

// RenderDevcontainer renders a dev container with the toolchains CI installs,
// at the same versions, and smith at the version the pipelines install; it
// then runs smith ci setup, as CI does.
func RenderDevcontainer(name string, m Model) string {
	features := map[string]map[string]string{}
	var ids []string
	for _, t := range m.Toolchains {
		if t.Devcontainer == nil {
			continue
		}
		if _, seen := features[t.Devcontainer.ID]; !seen {
			ids = append(ids, t.Devcontainer.ID)
		}
		opts := map[string]string{}
		for k, v := range t.Devcontainer.Options {
			opts[k] = v
		}
		features[t.Devcontainer.ID] = opts
	}
	sort.Strings(ids)
	doc := struct {
		Name     string                       `json:"name"`
		Image    string                       `json:"image"`
		Features map[string]map[string]string `json:"features"`
		Post     string                       `json:"postCreateCommand"`
	}{
		Name: name, Image: DevcontainerImage, Features: features,
		Post: devcontainerPost(m),
	}
	var b strings.Builder
	b.WriteString(header("//"))
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(doc)
	return b.String()
}

// devcontainerPost installs smith as the pipelines do, then runs smith ci
// setup; the install may not be on the container's PATH yet, so it is added
// for that command.
func devcontainerPost(m Model) string {
	install := SmithInstallFor(m)
	if install.Script {
		return `sh ` + InstallScriptPath + ` "` + m.SmithVersion + `" && "$HOME/.local/bin/smith" ci setup`
	}
	path := `"$HOME/.local/bin"`
	if dir := install.ContainerPathDir(); dir != "" {
		path = `"` + dir + `"`
	}
	return install.ContainerCommand() + ` && PATH=` + path + `:"$PATH" smith ci setup`
}
