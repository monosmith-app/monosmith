package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pelletier/go-toml/v2"
)

// ErrNotFound is returned when no smith.toml exists at or above a directory.
var ErrNotFound = errors.New("not inside a smith workspace (no smith.toml found)")

// FindRoot walks up from dir to the directory holding smith.toml.
func FindRoot(dir string) (string, error) {
	if env := os.Getenv("SMITH_WORKSPACE_ROOT"); env != "" {
		dir = env
	}
	return FindRootFrom(dir)
}

// FindRootFrom is FindRoot for a directory the caller chose: it ignores
// SMITH_WORKSPACE_ROOT, which smith sets for its own tasks and which would
// otherwise send a copy of a workspace back to the real one.
func FindRootFrom(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	start := dir
	for {
		if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", MissingWorkspace(start)
		}
		dir = parent
	}
}

// Load reads smith.toml under root, rejects unknown keys and applies defaults.
func Load(root string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, MissingWorkspace(root)
	}
	if err != nil {
		return Config{}, err
	}
	return Parse(raw)
}

// Parse decodes smith.toml text, rejecting unknown keys so a typo is an error
// rather than a silently ignored setting.
func Parse(raw []byte) (Config, error) {
	var cfg Config
	dec := toml.NewDecoder(bytes.NewReader(raw)).DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) {
			detail := fmt.Sprintf("%s: %s", FileName, strings.TrimSpace(strict.String()))
			if needsNewerSmith(raw, strict) {
				return cfg, &NewerSmithError{Detail: detail}
			}
			return cfg, errors.New(detail)
		}
		var decodeErr *toml.DecodeError
		if errors.As(err, &decodeErr) {
			row, col := decodeErr.Position()
			return cfg, fmt.Errorf("%s:%d:%d: %s", FileName, row, col, decodeErr.Error())
		}
		return cfg, fmt.Errorf("%s: %w", FileName, err)
	}
	// An absent max_size_mb takes the default; an explicit 0 means no limit.
	var probe struct {
		Cache struct {
			MaxSizeMB *int `toml:"max_size_mb"`
		} `toml:"cache"`
	}
	if toml.Unmarshal(raw, &probe) == nil && probe.Cache.MaxSizeMB != nil && *probe.Cache.MaxSizeMB == 0 {
		cfg.Cache.MaxSizeMB = -1
	}
	return cfg, Validate(&cfg)
}

// needsNewerSmith says whether the unknown keys of smith.toml come from a newer
// smith: the file declares a newer schema, has a table this smith lacks, or has
// more than one unknown key. One unknown key alone is taken for a typo.
func needsNewerSmith(raw []byte, strict *toml.StrictMissingError) bool {
	var probe struct {
		Schema int `toml:"schema"`
	}
	if toml.Unmarshal(raw, &probe) == nil && probe.Schema > SchemaVersion {
		return true
	}
	if len(strict.Errors) > 1 {
		return true
	}
	return strings.Contains(strict.String(), "missing table")
}

// Validate applies defaults and checks values.
func Validate(cfg *Config) error {
	if cfg.Schema == 0 {
		cfg.Schema = SchemaVersion
	}
	if cfg.Schema > SchemaVersion {
		return &NewerSmithError{Detail: fmt.Sprintf("%s uses schema %d; this smith reads up to %d, upgrade smith", FileName, cfg.Schema, SchemaVersion)}
	}
	if cfg.Schema < SchemaVersion {
		return fmt.Errorf("%s uses schema %d, this smith schema %d: run `smith upgrade`", FileName, cfg.Schema, SchemaVersion)
	}
	if err := validateDefaultsVintage(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Workspace.Name) == "" {
		return fmt.Errorf("%s: [workspace] name is required", FileName)
	}
	for _, glob := range cfg.Workspace.Exclude {
		if !doublestar.ValidatePattern(glob) || path.IsAbs(glob) || strings.Contains(glob, `\`) {
			return fmt.Errorf("%s: [workspace] exclude: %q is not a workspace-relative glob with forward slashes", FileName, glob)
		}
	}
	if r := cfg.Workspace.Recipe; r != "" && !ValidRecipeRef(r) {
		return fmt.Errorf("%s: [workspace] recipe: %q is not an https URL or a path relative to smith.toml with forward slashes", FileName, r)
	}
	for kind, r := range cfg.Recipes {
		if !slices.Contains(RecipeKinds, kind) {
			return fmt.Errorf("%s: [recipes] %s: the kind is app, lib or package", FileName, kind)
		}
		if !ValidRecipeRef(r) {
			return fmt.Errorf("%s: [recipes] %s: %q is not an https URL or a path relative to smith.toml with forward slashes", FileName, kind, r)
		}
	}
	for name, st := range cfg.Stacks {
		if strings.TrimSpace(st.Language) == "" && len(st.Choices) == 0 {
			return fmt.Errorf("%s: [stacks.%s] needs a language plugin (language = \"node\") or choices (choices = [\"react\"])", FileName, name)
		}
	}
	if len(cfg.CI.Providers) == 0 {
		cfg.CI.Providers = []string{"github"}
	}
	for _, p := range cfg.CI.Providers {
		if p != "github" && p != "azure" && p != "gitlab" && p != "circleci" && p != "bitbucket" {
			return fmt.Errorf("%s: unknown CI provider %q (use github, azure, gitlab, circleci or bitbucket)", FileName, p)
		}
	}
	if cfg.Release.RubyGemsTrustedPublishing && slices.ContainsFunc(cfg.CI.Providers, func(p string) bool { return p != "github" }) {
		return fmt.Errorf("%s: [release] rubygems_trusted_publishing works on GitHub Actions only: set [ci] providers = [\"github\"]", FileName)
	}
	if cfg.Release.VSCEEntra && slices.ContainsFunc(cfg.CI.Providers, func(p string) bool { return p != "github" }) {
		return fmt.Errorf("%s: [release] vsce_entra works on GitHub Actions only: set [ci] providers = [\"github\"]", FileName)
	}
	for _, list := range [][]string{cfg.CI.AuditPolicies, cfg.CI.AuditWarn} {
		for _, p := range list {
			if p != "node" && p != "go" && p != "python" && p != "rust" {
				return fmt.Errorf("%s: unknown audit policy %q (use node, go, python or rust)", FileName, p)
			}
		}
	}
	for _, b := range cfg.CI.Branches {
		if strings.TrimSpace(b) == "" {
			return fmt.Errorf("%s: [ci] branches must not hold an empty branch name", FileName)
		}
	}
	if err := validateSiblingRepos(cfg); err != nil {
		return err
	}
	switch cfg.CI.Attest {
	case "", "on", "optional", "off":
	default:
		return fmt.Errorf("%s: [ci] attest must be \"on\", \"optional\" or \"off\", not %q", FileName, cfg.CI.Attest)
	}
	if err := validateCISecrets(cfg.CI.Secrets); err != nil {
		return err
	}
	if cfg.CI.AuditMaxUnsafe < 0 {
		return fmt.Errorf("%s: audit_max_unsafe must not be negative", FileName)
	}
	if cfg.Release.Branch == "" {
		cfg.Release.Branch = "main"
	}
	if cfg.Release.Remote == "" {
		cfg.Release.Remote = "origin"
	}
	if err := validateReleaseChannels(cfg.Release); err != nil {
		return err
	}
	switch cfg.Release.APICheck {
	case "":
		cfg.Release.APICheck = "warn"
	case "warn", "error", "off":
	default:
		return fmt.Errorf("%s: [release] api_check must be \"warn\", \"error\" or \"off\", not %q", FileName, cfg.Release.APICheck)
	}
	if cfg.Cache.Dir == "" {
		cfg.Cache.Dir = ".smith/cache"
	}
	if cfg.Cache.MaxSizeMB == 0 {
		cfg.Cache.MaxSizeMB = 2048
	}
	for name, conv := range cfg.Conventions {
		if name != "docker" && (conv.Registry != "" || conv.Registries != nil || conv.DescriptionFrom != "" || conv.ShortDescription != "" || conv.TagCommit != nil || conv.TagBranch != nil || conv.Push != nil || conv.Verify != nil) {
			return fmt.Errorf("%s: [conventions.%s] registry, registries, description_from, short_description, tag_commit, tag_branch, push and verify are settings of the docker convention", FileName, name)
		}
		if name != "docker" {
			continue
		}
		label := "[conventions.docker]"
		registries, err := checkDockerRegistries(label, conv.Registry, conv.Registries, conv.DescriptionFrom)
		if err != nil {
			return fmt.Errorf("%s: %w", FileName, err)
		}
		conv.Registries = registries
		cfg.Conventions[name] = conv
	}
	if err := validateDeclaredProjects(cfg.Projects); err != nil {
		return err
	}
	for name, po := range cfg.Projects {
		if po.Docker != nil {
			if _, ok := cfg.Conventions["docker"]; !ok {
				return fmt.Errorf("%s: [projects.%s.docker] needs [conventions.docker]", FileName, name)
			}
			if c := po.Docker.Context; c != "" && !ValidContext(c) {
				return fmt.Errorf("%s: [projects.%s.docker] context %q must be a workspace-relative directory with forward slashes (\".\" or apps/%s, say)", FileName, name, c, name)
			}
			registries, err := checkDockerRegistries("[projects."+name+".docker]", po.Docker.Registry, po.Docker.Registries, po.Docker.DescriptionFrom)
			if err != nil {
				return fmt.Errorf("%s: %w", FileName, err)
			}
			po.Docker.Registries = registries
		}
		// The release table is checked and completed by ResolveRelease once the
		// project's plugins have had their say (a Tauri app's plugin supplies
		// build, assets and platforms).
	}
	if err := validateBareTags(cfg); err != nil {
		return err
	}
	if err := validateTagFormat(cfg); err != nil {
		return err
	}
	if err := validateBoundaries(cfg); err != nil {
		return err
	}
	if err := validateTaskGroups(cfg); err != nil {
		return err
	}
	if err := validateRegistries(cfg); err != nil {
		return err
	}
	for name, ext := range cfg.Plugins.External {
		if len(ext.Command) > 0 && strings.TrimSpace(ext.Command[0]) == "" {
			return fmt.Errorf("%s: [plugins.external.%s] command starts with an empty program", FileName, name)
		}
		if ext.SHA256 != "" && !pluginDigest.MatchString(ext.SHA256) {
			return fmt.Errorf("%s: [plugins.external.%s] sha256 must be 64 hexadecimal characters, optionally after \"sha256:\"", FileName, name)
		}
		for _, v := range ext.Env {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("%s: [plugins.external.%s] env has an empty name", FileName, name)
			}
		}
	}
	return nil
}

// pluginDigest is how a plugin's sha256 pin is written.
var pluginDigest = regexp.MustCompile(`^(?i:sha256:)?[0-9a-fA-F]{64}$`)
