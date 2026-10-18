package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	write("apps/api/go.mod", "module example.com/api\n\ngo 1.25.3\n")
	write("apps/web/index.html", "<!doctype html>")
	write("apps/worker/pyproject.toml", "")
	app := func(name string) protocol.GenerateParams {
		return protocol.GenerateParams{Kind: protocol.ProjectApp, Name: name, Directory: "apps/" + name}
	}
	res, err := Generate(ws, app("api"))
	if err != nil || len(res.Files) != 2 {
		t.Fatal(err, res)
	}
	if d := read("apps/api/Dockerfile"); !strings.Contains(d, "FROM golang:1.25-alpine AS build") || !strings.Contains(d, "go build -trimpath -o /out/app ./apps/api") {
		t.Errorf("Go Dockerfile:\n%s", d)
	}
	if !strings.Contains(read(".dockerignore"), "**/node_modules") {
		t.Error("the context leaves node_modules out")
	}
	if res, _ := Generate(ws, app("web")); len(res.Files) != 1 || !strings.Contains(read("apps/web/Dockerfile"), "COPY dist/apps/web/ /usr/share/nginx/html/") {
		t.Errorf("a web app is served by nginx; .dockerignore is written once: %+v", res)
	}
	if res, _ := Generate(ws, app("worker")); len(res.Files) != 0 || !strings.Contains(res.Notes[0], "Write apps/worker/Dockerfile") {
		t.Errorf("an app without a recipe gets a note: %+v", res)
	}
	write("apps/svc/package.json", `{"name":"@acme/svc","scripts":{"build":"tsc"}}`)
	if res, _ := Generate(ws, app("svc")); len(res.Files) != 1 {
		t.Errorf("a Node app that builds gets a Dockerfile: %+v", res)
	}
	d := read("apps/svc/Dockerfile")
	for _, want := range []string{"FROM node:24-slim", "COPY dist/apps/svc/package.json dist/apps/svc/package-lock.json ./", "COPY dist/apps/svc/workspace_modules ./workspace_modules", "RUN npm ci --omit=dev --ignore-scripts", `CMD ["node", "."]`} {
		if !strings.Contains(d, want) {
			t.Errorf("Node Dockerfile lacks %q:\n%s", want, d)
		}
	}
	write("apps/plain/package.json", `{"name":"@acme/plain"}`)
	if res, _ := Generate(ws, app("plain")); len(res.Files) != 0 {
		t.Errorf("a Node app with no build is the author's to ship: %+v", res)
	}
	write("apps/api/Dockerfile", "FROM scratch\n")
	if res, _ := Generate(ws, app("api")); len(res.Files) != 0 || read("apps/api/Dockerfile") != "FROM scratch\n" {
		t.Error("an existing Dockerfile is kept")
	}
	if res, _ := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core"}); len(res.Files) != 0 {
		t.Error("a lib ships no image")
	}
}
