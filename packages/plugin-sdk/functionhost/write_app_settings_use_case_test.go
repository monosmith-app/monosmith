package functionhost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAppSettingsTurnsEachEnvironmentIntoOneLineOfAppSettings(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "apps", "orders")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(app, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("local.settings.json", `{"Values": {"MODE": "laptop"}}`)
	write("local.settings.sample.json", `{"Values": {"MODE": "sample"}}`)
	write("local.settings.uat.json", `{"IsEncrypted": false, "Values": {
	  "MODE": "uat", "RETRIES": 3, "FEATURE_X": true, "ApiUrl": "https://api/uat?a=1&b=<2>",
	  "AzureWebJobsStorage": "UseDevelopmentStorage=true", "APPLICATIONINSIGHTS_CONNECTION_STRING": "x"
	}, "ConnectionStrings": {"Db": "Server=uat"}}`)
	write("local.settings.prd.json", `{"Values": {"MODE": "prd"}}`)

	written, notes, err := WriteAppSettings(app, root, "orders", []string{"MODE"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(written, ",") != "dist/drop/orders/env/prd.json,dist/drop/orders/env/uat.json" {
		t.Fatalf("one file per environment, local.settings.json and the sample left out: %v", written)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "dist", "drop", "orders", "env", "uat.json"))
	want := `[{"name":"ApiUrl","value":"https://api/uat?a=1&b=<2>","slotSetting":false},` +
		`{"name":"FEATURE_X","value":"true","slotSetting":false},` +
		`{"name":"MODE","value":"uat","slotSetting":true},` +
		`{"name":"RETRIES","value":"3","slotSetting":false}]`
	if string(raw) != want {
		t.Fatalf("one line, sorted, without the settings Azure sets, slot settings marked:\n%s\nwant\n%s", raw, want)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "local.settings.uat.json has ConnectionStrings") {
		t.Errorf("connection strings are left out with a word: %v", notes)
	}
}

func TestWriteAppSettingsWritesNothingWithoutEnvironmentFiles(t *testing.T) {
	root := t.TempDir()
	written, notes, err := WriteAppSettings(root, root, "api", nil)
	if err != nil || len(written) != 0 || len(notes) != 0 {
		t.Fatalf("an app without local.settings.<env>.json is not an Azure app with settings: %v %v %v", written, notes, err)
	}
	if err := os.WriteFile(filepath.Join(root, "local.settings.dev.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := WriteAppSettings(root, root, "api", nil); err == nil || !strings.Contains(err.Error(), "local.settings.dev.json is not valid JSON") {
		t.Fatalf("a broken file is named: %v", err)
	}
}
