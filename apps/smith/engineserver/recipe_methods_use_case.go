package engineserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
	smithworkspace "github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// recipeList answers with the recipe files of the workspace.
func (s *Server) recipeList(_ context.Context, _ json.RawMessage) (any, *rpcError) {
	root, err := smithworkspace.FindRoot(s.dir())
	if err != nil {
		return nil, &rpcError{codeFailed, err.Error()}
	}
	listed, err := recipecatalog.ListRecipes(root)
	if err != nil {
		return nil, &rpcError{codeFailed, err.Error()}
	}
	return RecipeList{Recipes: listed}, nil
}

// loadRecipe reads the recipe a method names: a file inside the workspace, or
// an https URL (pinned in the workspace's lockfile like for smith recipe check).
func (s *Server) loadRecipe(ref string) (recipe.Loaded, *rpcError) {
	root, err := smithworkspace.FindRoot(s.dir())
	if err != nil {
		return recipe.Loaded{}, &rpcError{codeFailed, err.Error()}
	}
	if ref == "" || strings.HasPrefix(ref, "-") {
		return recipe.Loaded{}, &rpcError{codeInvalidParams, `"recipe" must name a recipe file in the workspace, or an https URL`}
	}
	path := ref
	if !strings.Contains(ref, "://") {
		abs := filepath.Join(root, filepath.FromSlash(ref))
		if inside, relErr := filepath.Rel(root, abs); relErr != nil || inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) || filepath.IsAbs(ref) {
			return recipe.Loaded{}, &rpcError{codeInvalidParams, "a recipe file must be inside the workspace: " + ref}
		}
		path = abs
	}
	loaded, err := (&recipe.Loader{Root: root}).Load(path)
	if err != nil {
		return recipe.Loaded{}, &rpcError{codeFailed, err.Error()}
	}
	return loaded, nil
}

// recipeDescribe answers with what a recipe offers before any choice is
// answered: its choices, the levels and slots it fills, the plugins it
// configures.
func (s *Server) recipeDescribe(_ context.Context, raw json.RawMessage) (any, *rpcError) {
	var p RecipeParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	loaded, rerr := s.loadRecipe(p.Recipe)
	if rerr != nil {
		return nil, rerr
	}
	return recipecatalog.Describe(loaded), nil
}

// recipeRoles answers with what a recipe puts in each role (a level such as
// language or deploy, or a slot such as lint) once its choices are answered:
// the plugins, and the file and line each came from.
func (s *Server) recipeRoles(_ context.Context, raw json.RawMessage) (any, *rpcError) {
	var p RecipeParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	loaded, rerr := s.loadRecipe(p.Recipe)
	if rerr != nil {
		return nil, rerr
	}
	resolved, err := recipe.Resolve(loaded, p.Choices)
	if err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	return RecipeRoles{Name: resolved.Name, Version: resolved.Version, Choices: resolved.Asked, Roles: resolved.Uses}, nil
}
