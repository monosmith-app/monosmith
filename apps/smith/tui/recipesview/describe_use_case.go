package recipesview

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ReadCatalog lists the recipe files of the workspace at root and the recipe
// each kind of project is made from, as smith.toml says.
func ReadCatalog(root string) Catalog {
	cfg, err := workspace.Load(root)
	if err != nil {
		return Catalog{Err: err}
	}
	listed, err := recipecatalog.ListRecipes(root)
	if err != nil {
		return Catalog{Err: err}
	}
	byKind := map[string]string{}
	for _, kind := range Kinds {
		byKind[kind] = workspace.RecipeFor(cfg, kind)
	}
	return Catalog{Recipes: listed, ByKind: byKind}
}

// Describe loads the recipe ref (a file relative to the workspace root, or an
// https URL) and resolves it for the answers given, without writing anything:
// the same work `smith add` does before it creates a project. Answers to
// choices the recipe does not ask for these answers (an earlier answer that a
// later one made irrelevant) are left out.
func Describe(root, ref string, answers map[string]string) Detail {
	d := Detail{Ref: ref}
	if ref == "" {
		return d
	}
	path := ref
	if !strings.Contains(ref, "://") {
		path = filepath.Join(root, filepath.FromSlash(ref))
	}
	loaded, err := (&recipe.Loader{Root: root}).Load(path)
	if err != nil {
		d.Problem = err.Error()
		return d
	}
	for _, o := range recipecatalog.Describe(loaded).Uses {
		d.Offers = append(d.Offers, Offer{Key: o.Key, Plugins: o.Plugins, Alternatives: o.Alternatives})
	}
	res, err := recipe.Resolve(loaded, answers)
	if err != nil && len(res.Asked) > 0 {
		// An answer the new choices made irrelevant is an error to the resolver; drop it and look again.
		kept := map[string]string{}
		for _, q := range res.Asked {
			if a, ok := answers[q.Name]; ok {
				kept[q.Name] = a
			}
		}
		if again, err2 := recipe.Resolve(loaded, kept); err2 == nil || isMissing(err2) {
			res, err = again, err2
		}
	}
	d.Name, d.Version = res.Name, res.Version
	for _, q := range res.Asked {
		d.Questions = append(d.Questions, Question{Name: q.Name, Ask: q.Ask, Options: q.Options, Default: q.Default, Answer: q.Answer, Given: q.Given})
	}
	var missing *recipe.MissingAnswerError
	switch {
	case errors.As(err, &missing):
		for _, q := range missing.Choices {
			d.Unanswered = append(d.Unanswered, q.Name)
		}
		return d
	case err != nil:
		d.Problem = err.Error()
		return d
	}
	for _, u := range res.Uses {
		d.Roles = append(d.Roles, Role{Key: u.Key, Plugins: u.Plugins, Origin: origin(u.Origin, root)})
	}
	for plugin := range res.Settings {
		d.Configures = append(d.Configures, plugin)
	}
	sort.Strings(d.Configures)
	return d
}

func isMissing(err error) bool {
	var missing *recipe.MissingAnswerError
	return errors.As(err, &missing)
}

// origin is where a value came from, with a file inside the workspace shown relative to it.
func origin(o recipe.Origin, root string) string {
	if filepath.IsAbs(o.File) {
		if rel, err := filepath.Rel(root, o.File); err == nil && !strings.HasPrefix(rel, "..") {
			o.File = filepath.ToSlash(rel)
		}
	}
	return o.String()
}
