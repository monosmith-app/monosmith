package functions

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorWarnsWhenTheLocalToolOfAHostIsMissing(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		path := filepath.Join(ws, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("apps/az/package.json", `{"dependencies":{"@azure/functions":"^4"}}`)
	write("apps/az2/package.json", `{"dependencies":{"@azure/functions":"^4"}}`)
	write("apps/aw/package.json", `{}`)
	write("apps/aw/template.yaml", "Transform: AWS::Serverless-2016-10-31\n")
	write("apps/web/package.json", `{}`)
	real := lookPath
	defer func() { lookPath = real }()
	lookPath = func(name string) (string, error) {
		if name == "sam" {
			return "/bin/sam", nil
		}
		return "", errors.New("not found")
	}
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Name: "az", Root: "apps/az"}, {Name: "az2", Root: "apps/az2"}, {Name: "aw", Root: "apps/aw"}, {Name: "web", Root: "apps/web"}}})
	if len(res.Findings) != 2 {
		t.Fatalf("one finding per host in use: %+v", res.Findings)
	}
	for _, f := range res.Findings {
		wantOK := f.Check == "the AWS SAM CLI is installed"
		if f.OK != wantOK || !f.Warning || (!f.OK && f.Remedy == "") {
			t.Errorf("%+v", f)
		}
	}
}
