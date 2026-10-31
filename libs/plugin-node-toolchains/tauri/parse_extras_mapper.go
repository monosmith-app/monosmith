package tauri

import (
	"fmt"
	"slices"
	"strings"
)

// ParseExtras reads a project's [projects.<p>.settings.tauri] table. A key it
// does not know is an error, so a typo does not silently bundle nothing.
func ParseExtras(settings map[string]any) (Extras, error) {
	var e Extras
	for key, v := range settings {
		if key != "sidecars" && key != "resources" {
			return e, fmt.Errorf("settings.tauri: unknown key %q (known: sidecars, resources)", key)
		}
		list, ok := v.([]any)
		if !ok {
			return e, fmt.Errorf("settings.tauri.%s must be a list", key)
		}
		for i, item := range list {
			var err error
			if key == "sidecars" {
				err = e.addSidecar(item)
			} else {
				err = e.addResource(item)
			}
			if err != nil {
				return e, fmt.Errorf("settings.tauri.%s[%d]: %w", key, i, err)
			}
		}
	}
	return e, nil
}

func (e *Extras) addSidecar(item any) error {
	t, err := textTable(item, "name", "project", "bin", "from")
	if err != nil {
		return err
	}
	s := Sidecar{Name: t["name"], Project: t["project"], Bin: t["bin"], From: t["from"]}
	switch {
	case s.Name == "" || strings.ContainsAny(s.Name, `/\`) || strings.HasSuffix(s.Name, ".exe"):
		return fmt.Errorf("name is the program's file name without a folder or .exe, such as \"videocraw\"")
	case (s.Project == "") == (s.From == ""):
		return fmt.Errorf("%s: give exactly one of project (a project of the workspace) and from (a file)", s.Name)
	case s.Bin != "" && s.Project == "":
		return fmt.Errorf("%s: bin goes with project", s.Name)
	}
	e.Sidecars = append(e.Sidecars, s)
	return nil
}

func (e *Extras) addResource(item any) error {
	if from, ok := item.(string); ok {
		item = map[string]any{"from": from}
	}
	t, err := textTable(item, "from", "to")
	if err != nil {
		return err
	}
	if t["from"] == "" {
		return fmt.Errorf("from is required: a path, a folder or a pattern, relative to the workspace (environment variables as $NAME)")
	}
	e.Resources = append(e.Resources, Resource{From: t["from"], To: t["to"]})
	return nil
}

// textTable reads a table of text values that has only the keys given.
func textTable(item any, keys ...string) (map[string]string, error) {
	m, ok := item.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("must be a table such as { name = \"x\", project = \"y\" }")
	}
	out := map[string]string{}
	for k, v := range m {
		s, isText := v.(string)
		if !slices.Contains(keys, k) || !isText {
			return nil, fmt.Errorf("%q is not a text key of this table (keys: %s)", k, strings.Join(keys, ", "))
		}
		out[k] = s
	}
	return out, nil
}
