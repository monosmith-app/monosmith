package recipestack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Declare appends a declaration to smith.toml, keeping everything already in
// it, and checks that the result still loads. It returns a function that puts
// smith.toml back as it was, for a project whose creation failed after this.
func Declare(root string, d Declaration) (undo func() error, err error) {
	path := filepath.Join(root, workspace.FileName)
	original, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(original)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "\n" + Render(d)
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return nil, fmt.Errorf("the declaration for %s would break smith.toml: %w", d.Project, err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return nil, err
	}
	return func() error { return os.WriteFile(path, original, 0o644) }, nil
}
