package releasing

import (
	"errors"
	"fmt"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

func TestPublished(t *testing.T) {
	p := protocol.PublishedParams{Project: protocol.Project{Exports: []string{"@acme/sdk"}}}
	for _, c := range []struct {
		out, errOut string
		err         error
		want        protocol.PublishedResult
	}{
		{"1.4.0\n", "", nil, protocol.PublishedResult{Version: "1.4.0"}},
		{"", "npm error code E404\nnpm error 404 Not Found - GET https://registry/@acme%2fsdk", errors.New("exit 1"), protocol.PublishedResult{}},
		{"", "npm error code ECONNREFUSED\nnpm error connect ECONNREFUSED", errors.New("exit 1"), protocol.PublishedResult{Unreachable: true, Detail: "npm error code ECONNREFUSED"}},
		{"", "", fmt.Errorf("npm view %w after 8s", toolrun.ErrTimeout), protocol.PublishedResult{Unreachable: true, Detail: protocol.RegistryUnreachableDetail}},
	} {
		capture = func(_ string, argv []string) (string, string, error) {
			if argv[1] == "config" {
				return "https://registry.example/\n", "", nil
			}
			return c.out, c.errOut, c.err
		}
		got, err := Published(t.TempDir(), p)
		c.want.Registry = "https://registry.example/"
		if err != nil || got != c.want {
			t.Errorf("got %+v %v, want %+v", got, err, c.want)
		}
	}
}

func TestPublishedAsksNpmForTheRegistryOnceForManyProjects(t *testing.T) {
	registries.forget()
	real := capture
	defer func() { capture = real }()
	configs := 0
	capture = func(_ string, argv []string) (string, string, error) {
		if argv[1] == "config" {
			configs++
			return "https://registry.example/\n", "", nil
		}
		return "1.0.0\n", "", nil
	}
	root := t.TempDir()
	for _, name := range []string{"@acme/a", "@acme/b", "@acme/c"} {
		if _, err := Published(root, protocol.PublishedParams{Project: protocol.Project{Exports: []string{name}}}); err != nil {
			t.Fatal(err)
		}
	}
	if configs != 1 {
		t.Fatalf("npm config get registry ran %d times, want 1", configs)
	}
}
