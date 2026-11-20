package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckNestedLockfileOffersToDeleteIt(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "package.json", `{"private":true,"workspaces":["apps/*"]}`)
	writeFile(t, ws, "apps/api/package.json", `{"name":"api"}`)
	writeFile(t, ws, "apps/api/package-lock.json", `{}`)
	writeFile(t, ws, "apps/web/package.json", `{"name":"web"}`)

	f, ok := CheckNestedLockfile(ws, protocol.Project{Name: "api", Root: "apps/api"})
	if !ok || !f.Warning || f.Fix == nil || len(f.Fix.Files) != 1 || !f.Fix.Files[0].Delete || f.Fix.Files[0].Path != "apps/api/package-lock.json" {
		t.Fatalf("finding %+v", f)
	}
	if f.Fix.Safe {
		t.Fatal("deleting a file of the user's is asked first")
	}
	if !strings.Contains(f.Detail, "root package-lock.json") {
		t.Fatalf("detail %q", f.Detail)
	}
	if _, ok := CheckNestedLockfile(ws, protocol.Project{Name: "web", Root: "apps/web"}); ok {
		t.Fatal("a project without a lockfile is fine")
	}

	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}}, "")
	found := false
	for _, f := range res.Findings {
		found = found || f.Check == "apps/api has no lockfile of its own"
	}
	if !found {
		t.Fatal("doctor reports the nested lockfile of a covered project")
	}
}
