package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAnNxOrTurborepoFolderIsConvertible(t *testing.T) {
	for file, want := range map[string]bool{"nx.json": true, "turbo.json": true, "package.json": false} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, file), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := Convertible(dir); got != want {
			t.Errorf("%s: Convertible = %v, want %v", file, got, want)
		}
	}
	dir := t.TempDir()
	for _, f := range []string{"nx.json", "smith.toml"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if Convertible(dir) {
		t.Error("a smith workspace is not convertible")
	}
}

// handedOver runs the hand-over against a stand-in convert and returns the
// flags convert saw.
func handedOver(t *testing.T, level string, initialise, tags bool) (string, error) {
	t.Helper()
	var got []string
	root := &cobra.Command{Use: "smith"}
	conv := &cobra.Command{Use: "convert", RunE: func(c *cobra.Command, _ []string) error {
		for _, name := range []string{"dry-run", "level", "json"} {
			got = append(got, name+"="+c.Flags().Lookup(name).Value.String())
		}
		return nil
	}}
	conv.Flags().Bool("dry-run", false, "")
	conv.Flags().Bool("json", false, "")
	conv.Flags().String("level", "conservative", "")
	adopt := Command("test")
	root.AddCommand(conv, adopt)
	err := handOverToConvert(adopt, level, false, false, initialise, tags, false, false, false)
	return strings.Join(got, " "), err
}

func TestAdoptHandsAnNxWorkspaceToConvert(t *testing.T) {
	if got, err := handedOver(t, "full", false, false); err != nil || got != "dry-run=true level=full json=false" {
		t.Errorf("a report is convert --dry-run: %q, %v", got, err)
	}
	if got, err := handedOver(t, "conservative", true, false); err != nil || got != "dry-run=false level=conservative json=false" {
		t.Errorf("--init applies: %q, %v", got, err)
	}
	if _, err := handedOver(t, "conservative", true, true); err == nil || !strings.Contains(err.Error(), "--tags") {
		t.Errorf("--tags is not for a conversion: %v", err)
	}
}
