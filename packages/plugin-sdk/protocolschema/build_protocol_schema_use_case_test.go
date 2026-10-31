package protocolschema

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const (
	protocolDir   = "../protocol"
	committedPath = "../schema/protocol.schema.json"
)

// The drift guard: the committed schema is what the contract of every other
// language's SDK is generated from, so it must be what the Go types say.
func TestCommittedSchemaIsCurrent(t *testing.T) {
	want, err := Build(protocolDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("%v\nrun `go run ./schemagen` in packages/plugin-sdk to write it", err)
	}
	if !bytes.Equal(bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n")), want) {
		t.Fatal("schema/protocol.schema.json is not what the protocol package describes: run `go run ./schemagen` in packages/plugin-sdk and commit the result (then regenerate the SDKs' types)")
	}
}

type builtSchema struct {
	Version   int                        `json:"x-protocolVersion"`
	Methods   map[string]json.RawMessage `json:"x-methods"`
	Codes     []struct{ Name string }    `json:"x-errorCodes"`
	Roles     []struct{ Name string }    `json:"x-roles"`
	Transport map[string]any             `json:"x-transport"`
	Defs      map[string]struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	} `json:"$defs"`
}

func buildAndParse(t *testing.T) (builtSchema, string) {
	t.Helper()
	raw, err := Build(protocolDir)
	if err != nil {
		t.Fatal(err)
	}
	var s builtSchema
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("the schema is not JSON: %v", err)
	}
	return s, string(raw)
}

func TestSchemaDescribesEveryMethodAndResolvesEveryReference(t *testing.T) {
	s, raw := buildAndParse(t)
	if s.Version != protocol.Version {
		t.Errorf("x-protocolVersion = %d, the protocol speaks %d", s.Version, protocol.Version)
	}
	if len(s.Methods) != 21 {
		t.Errorf("want the 21 methods of version 1, got %d", len(s.Methods))
	}
	for name, row := range s.Methods {
		var m struct {
			Params, Result map[string]string
		}
		if err := json.Unmarshal(row, &m); err != nil {
			t.Fatal(err)
		}
		for _, r := range []map[string]string{m.Params, m.Result} {
			if r == nil {
				if name != protocol.MethodCancel {
					t.Errorf("method %s has no params or result", name)
				}
				continue // a notification has no result
			}
			def := strings.TrimPrefix(r["$ref"], "#/$defs/")
			if _, ok := s.Defs[def]; !ok {
				t.Errorf("method %s refers to the undefined %q", name, r["$ref"])
			}
		}
	}
	// Every $ref in the file resolves.
	for _, part := range strings.Split(raw, `"$ref": "#/$defs/`)[1:] {
		def, _, _ := strings.Cut(part, `"`)
		if _, ok := s.Defs[def]; !ok {
			t.Errorf("$ref to the undefined definition %q", def)
		}
	}
	if len(s.Codes) != 7 || len(s.Roles) != 5 {
		t.Errorf("error codes %d, roles %d", len(s.Codes), len(s.Roles))
	}
	if s.Transport["maxMessageBytes"] != float64(64<<20) {
		t.Errorf("transport: %v", s.Transport)
	}
}

func TestOmitemptyDecidesWhatIsRequired(t *testing.T) {
	s, _ := buildAndParse(t)
	project := s.Defs["Project"]
	if got := strings.Join(project.Required, ","); got != "name,root,language,kind" {
		t.Errorf("Project requires %q", got)
	}
	if _, ok := project.Properties["publishTo"]; !ok {
		t.Error("Project has no publishTo")
	}
	if got := strings.Join(s.Defs["Step"].Required, ","); got != "name,command,blocking" {
		t.Errorf("Step requires %q: a bool without omitempty is required", got)
	}
}

func TestCancelIsTheOnlyNotificationAndHasNoResult(t *testing.T) {
	s, _ := buildAndParse(t)
	for name, row := range s.Methods {
		var m struct {
			Result       json.RawMessage `json:"result"`
			Capability   *string         `json:"capability"`
			Notification bool            `json:"notification"`
		}
		if err := json.Unmarshal(row, &m); err != nil {
			t.Fatal(err)
		}
		if m.Notification != (name == protocol.MethodCancel) {
			t.Errorf("%s: notification = %v", name, m.Notification)
		}
		if name == protocol.MethodCancel && (string(m.Result) != "null" || m.Capability == nil || *m.Capability != "cancel") {
			t.Errorf("cancel row: result %s, capability %v", m.Result, m.Capability)
		}
	}
}
