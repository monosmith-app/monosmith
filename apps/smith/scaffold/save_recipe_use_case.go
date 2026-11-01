package scaffold

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// askSaveRecipe offers to keep the answers as a recipe, and returns the name
// the person gives it, or "" for no. A recipe needs a main language; without
// one it does not ask and prints nothing. Any other answer than y is no.
func askSaveRecipe(opts NewOptions, in *bufio.Reader, out io.Writer) string {
	if _, err := renderRecipe(opts, opts.Name); err != nil {
		// Nobody asked for a recipe, so there is nothing to explain: the offer
		// is only made when these answers can become one.
		return ""
	}
	fmt.Fprint(out, "Save these answers as a recipe to reuse on other workspaces? [y/N]: ")
	line, _ := in.ReadString('\n')
	if answer := strings.ToLower(strings.TrimSpace(line)); answer != "y" && answer != "yes" {
		return ""
	}
	for attempt := 0; attempt < maxAnswerTries; attempt++ {
		fmt.Fprintf(out, "Name of the recipe [%s]: ", opts.Name)
		line, err := in.ReadString('\n')
		name := strings.TrimSpace(line)
		if name == "" {
			name = opts.Name
		}
		if problem := ValidateName("recipe", name); problem != nil {
			fmt.Fprintf(out, "%v\n", problem)
			if err != nil {
				break
			}
			continue
		}
		return name
	}
	fmt.Fprintln(out, "Not saving a recipe.")
	return ""
}

// recipeFiles are the recipe files a new workspace gets, by their path in it:
// the one saved from the answers (opts.SaveRecipe) and the one copied in from
// a file (opts.RecipeFile, which also becomes [workspace] recipe). It refuses
// a file that is already there, before anything is created.
func recipeFiles(dir string, opts *NewOptions) (map[string]string, error) {
	files := map[string]string{}
	if opts.SaveRecipe != "" {
		if err := ValidateName("recipe", opts.SaveRecipe); err != nil {
			return nil, fmt.Errorf("--save-recipe: %w", err)
		}
		text, err := renderRecipe(*opts, opts.SaveRecipe)
		if err != nil {
			return nil, fmt.Errorf("--save-recipe %s: %w", opts.SaveRecipe, err)
		}
		files[recipePath(opts.SaveRecipe)] = text
	}
	if opts.RecipeFile != "" {
		if opts.Recipe != "" {
			return nil, errors.New("use --recipe or --recipe-file, not both")
		}
		body, err := os.ReadFile(opts.RecipeFile)
		if err != nil {
			return nil, fmt.Errorf("--recipe-file: %w", err)
		}
		if _, err := (&recipe.Loader{}).Load(opts.RecipeFile); err != nil {
			return nil, fmt.Errorf("--recipe-file %s: %w", opts.RecipeFile, err)
		}
		path := "recipes/" + filepath.Base(opts.RecipeFile)
		files[path] = string(body)
		opts.Recipe = path
	}
	for path := range files {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(path))); err == nil {
			return nil, fmt.Errorf("%s already exists; choose another recipe name", path)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return files, nil
}

// writeRecipeFiles writes the files recipeFiles listed and returns their paths.
func writeRecipeFiles(dir string, files map[string]string, out io.Writer) ([]string, error) {
	var written []string
	for _, path := range sortedKeys(files) {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(full, []byte(files[path]), 0o644); err != nil {
			return nil, err
		}
		fmt.Fprintf(out, "  wrote %s\n", path)
		written = append(written, path)
	}
	return written, nil
}

// recipeAnswers loads the recipe a new workspace is created from and returns
// the answers it settles, so those questions are not asked. A recipe that
// cannot be read settles nothing here; NewWorkspace reports it.
func recipeAnswers(opts NewOptions) map[string]string {
	location := ""
	switch {
	case opts.RecipeFile != "":
		location = opts.RecipeFile
	case strings.HasPrefix(opts.Recipe, "https://"):
		location = opts.Recipe
	case opts.Recipe != "":
		location = filepath.Join(orDefault(opts.Dir, opts.Name), filepath.FromSlash(opts.Recipe))
	}
	if location == "" {
		return nil
	}
	loaded, err := (&recipe.Loader{}).Load(location)
	if err != nil {
		return nil
	}
	return answersFromRecipe(loaded)
}

// settledNames say, for a person, what the recipe answered.
var settledNames = []struct{ key, name string }{
	{AnswerLanguage, "the main language"},
	{AnswerNodeStack, "the Node tests and lint"},
	{AnswerArchitecture, "the code layout"},
}

// describeSettled says which questions the recipe answered, so the person
// knows why they are not asked.
func describeSettled(settled map[string]string) string {
	var parts []string
	for _, n := range settledNames {
		if _, ok := settled[n.key]; ok {
			parts = append(parts, n.name)
		}
	}
	return "The recipe already answers " + strings.Join(parts, ", ") + "."
}

// recipeSavedText closes smith new when a recipe was saved: where it is and
// how to use it for the next workspace.
func recipeSavedText(dir, name string) string {
	path := filepath.Join(dir, filepath.FromSlash(recipePath(name)))
	return "\nRecipe saved: " + path + "\nFor another workspace: smith new <name> --recipe-file " + path + "\n"
}
