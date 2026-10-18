package recipewizard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// ErrExists is returned when the recipe file is already there and the caller
// did not allow replacing it.
var ErrExists = errors.New("the recipe file exists")

// Creator turns a plan into a recipe file that is known to resolve.
type Creator struct {
	// Loader loads the parent recipe, if the plan has one.
	Loader *recipe.Loader
	// Lookup describes the plugins the recipe names.
	Lookup recipe.PluginLookup
}

// IsURL reports whether a reference to a recipe is an https URL.
func IsURL(ref string) bool { return strings.HasPrefix(ref, "https://") }

// ParentRef is where a plan's parent is, for a recipe written at target: a
// path in extends is relative to the recipe file, which is not yet written.
func ParentRef(target, extends string) string { return refIn(filepath.Dir(target), extends) }

// refIn is a reference to a recipe, made from a folder.
func refIn(dir, extends string) string {
	if extends == "" || IsURL(extends) || filepath.IsAbs(extends) {
		return extends
	}
	return filepath.Join(dir, filepath.FromSlash(extends))
}

// Render checks a plan and writes its recipe, then resolves the recipe with
// its parents for every combination of answers. It returns the text only
// when every branch resolves, so a recipe the wizard writes always does.
func (c Creator) Render(p Plan, target string) (string, error) {
	if err := p.Check(); err != nil {
		return "", err
	}
	text := Render(p)
	doc, err := recipe.ParseRecipe([]byte(text), recipe.Source{Ref: target})
	if err != nil {
		return "", fmt.Errorf("the recipe the plan makes does not parse (this is a bug in smith): %w", err)
	}
	var chain []recipe.Document
	if p.Extends != "" {
		parent, err := c.Loader.Load(ParentRef(target, p.Extends))
		if err != nil {
			return "", fmt.Errorf("extends %q: %w", p.Extends, err)
		}
		chain = parent.Chain
	}
	if err := recipe.ValidateAll(recipe.Loaded{Chain: append(chain, doc)}, c.Lookup); err != nil {
		return "", err
	}
	return text, nil
}

// Write puts the recipe at target, creating its folder. An existing file is
// replaced only when overwrite is set; otherwise the error is ErrExists.
func Write(target, text string, overwrite bool) error {
	if _, err := os.Stat(target); err == nil && !overwrite {
		return fmt.Errorf("%w: %s", ErrExists, target)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(text), 0o644)
}
