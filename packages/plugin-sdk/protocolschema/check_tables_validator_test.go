package protocolschema

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTablesMatchTheProtocolPackage(t *testing.T) {
	docs, err := ReadSourceDocs(protocolDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckTables(docs); err != nil {
		t.Fatal(err)
	}
}

// A method added to the protocol without a row fails the build with the names
// of everything that is out of step.
func TestAnUnknownMethodAndCodeAreReported(t *testing.T) {
	dir := t.TempDir()
	source := "package protocol\n\nconst MethodExtra = \"extra\"\n\nconst CodeExtra = -1\n"
	if err := os.WriteFile(filepath.Join(dir, "extra.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	docs, err := ReadSourceDocs(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = CheckTables(docs)
	if err == nil {
		t.Fatal("a protocol without the methods of the table is accepted")
	}
	for _, want := range []string{`MethodExtra ("extra") is missing`, "CodeExtra is missing", `the methods table has "initialize"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("not reported: %s\n%v", want, err)
		}
	}
}

// The capability a row names is the one the host checks before it calls the
// method: the table cannot say something the Go host does not do.
func TestCapabilityOfEachMethodIsWhatTheHostChecks(t *testing.T) {
	for _, spec := range Methods {
		none := &host.Client{Info: protocol.InitializeResult{}}
		if spec.Capability == "" {
			if !none.Supports(spec.Method) {
				t.Errorf("%s has no capability but the host does not always allow it", spec.Method)
			}
			continue
		}
		if none.Supports(spec.Method) {
			t.Errorf("%s is allowed with no capability", spec.Method)
		}
		var caps protocol.Capabilities
		reflect.ValueOf(&caps).Elem().FieldByName(spec.Capability).SetBool(true)
		withIt := &host.Client{Info: protocol.InitializeResult{Capabilities: caps}}
		if !withIt.Supports(spec.Method) {
			t.Errorf("%s is not allowed by the capability %s the table names", spec.Method, spec.Capability)
		}
	}
}
