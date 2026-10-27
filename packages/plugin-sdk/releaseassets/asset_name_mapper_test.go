package releaseassets

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAssetNameFillsPlaceholdersAndRejectsUnknownOnes(t *testing.T) {
	spec := protocol.AssetSpec{Product: "acme", OSAlias: map[string]string{"darwin": "macos"}}
	got, err := AssetName(spec, "cli", "1.0.0", "darwin", "arm64", "zip")
	if err != nil || got != "acme_1.0.0_macos_arm64.zip" {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := AssetName(protocol.AssetSpec{Name: "{nope}.zip"}, "cli", "1", "linux", "amd64", "zip"); err == nil {
		t.Error("an unknown placeholder is an error")
	}
}
