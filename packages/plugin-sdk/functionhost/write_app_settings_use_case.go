package functionhost

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// setByAzure are the settings Azure sets when it creates a function app: a
// deployment that sends them would override the app's own.
var setByAzure = map[string]bool{"APPLICATIONINSIGHTS_CONNECTION_STRING": true, "AzureWebJobsStorage": true}

// appSetting is one entry of the JSON that `az functionapp config appsettings
// set --settings @file` and the Azure portal's advanced edit read.
type appSetting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	SlotSetting bool   `json:"slotSetting"`
}

// WriteAppSettings turns each local.settings.<env>.json of an Azure Functions
// app (the Core Tools file, one per environment; local.settings.json itself is
// the developer's and stays out, like the sample a developer copies) into the app settings of that environment:
// one line of JSON, [{"name":..,"value":..,"slotSetting":..}], written to
// <workspace>/dist/drop/<app>/env/<env>.json, which `smith ci pack` keeps. A
// folder without such a file writes nothing, so every host can call it.
//
// Parameters: projectDir is the app's folder; workspaceRoot the workspace's;
// app names the app in dist/drop; slotSettings are the setting names that stay
// with a deployment slot (smith.toml [projects.<app>] slot_settings).
// Returns: the files written (workspace-relative), notes for the person, and an
// error when a settings file cannot be read or the output cannot be written.
func WriteAppSettings(projectDir, workspaceRoot, app string, slotSettings []string) (written, notes []string, err error) {
	files, err := filepath.Glob(filepath.Join(projectDir, "local.settings.*.json"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(files)
	for _, file := range files {
		env := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "local.settings."), ".json")
		if env == "sample" || env == "example" {
			continue // the template a developer copies, not an environment
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			return written, notes, err
		}
		var doc struct {
			Values            map[string]json.RawMessage `json:"Values"`
			ConnectionStrings map[string]json.RawMessage `json:"ConnectionStrings"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return written, notes, &settingsError{file: filepath.Base(file), err: err}
		}
		settings := make([]appSetting, 0, len(doc.Values))
		for name, value := range doc.Values {
			if setByAzure[name] {
				continue
			}
			settings = append(settings, appSetting{Name: name, Value: plain(value), SlotSetting: slices.Contains(slotSettings, name)})
		}
		sort.Slice(settings, func(i, j int) bool { return settings[i].Name < settings[j].Name })
		var out bytes.Buffer
		enc := json.NewEncoder(&out)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(settings); err != nil {
			return written, notes, err
		}
		rel := filepath.ToSlash(filepath.Join("dist", "drop", app, "env", env+".json"))
		target := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return written, notes, err
		}
		if err := os.WriteFile(target, bytes.TrimRight(out.Bytes(), "\n"), 0o644); err != nil {
			return written, notes, err
		}
		written = append(written, rel)
		if len(doc.ConnectionStrings) > 0 {
			notes = append(notes, filepath.Base(file)+" has ConnectionStrings, which are not app settings: set them as connection strings of the app in Azure")
		}
	}
	return written, notes, nil
}

// plain is a setting's value as Azure keeps it: a JSON string as it is, any
// other value (a number, a boolean) as its JSON text.
func plain(value json.RawMessage) string {
	var s string
	if json.Unmarshal(value, &s) == nil {
		return s
	}
	return string(bytes.TrimSpace(value))
}

// settingsError names the settings file that is not valid JSON.
type settingsError struct {
	file string
	err  error
}

// Error returns the sentence for the person.
func (e *settingsError) Error() string { return e.file + " is not valid JSON: " + e.err.Error() }

// Unwrap returns the JSON error.
func (e *settingsError) Unwrap() error { return e.err }
