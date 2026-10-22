package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListReadsRuntimePackagesOfTheLock(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "composer.lock"), []byte(`{"packages":[{"name":"monolog/monolog","version":"3.5.0"}],"packages-dev":[{"name":"phpunit/phpunit","version":"11.0.0"}]}`), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].PURL != "pkg:composer/monolog/monolog@3.5.0" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToComposerJSON(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"require":{"php":">=8.2","ext-json":"*","guzzlehttp/guzzle":"^7.8"}}`), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].Name != "guzzlehttp/guzzle" {
		t.Fatalf("%+v %v", res, err)
	}
}
