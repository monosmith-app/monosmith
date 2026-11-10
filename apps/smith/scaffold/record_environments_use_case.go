package scaffold

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// EnvironmentsTable is the smith.toml table that holds the environment settings
// of a web app: [projects.<name>.settings.vite].
func EnvironmentsTable(project string) string {
	return "projects." + configedit.TableKey(project) + ".settings." + webenv.PluginName
}

// EnvironmentsToml is smith.toml with the environments of a web app recorded:
// the list in the app's own table, and in [workspace] environments when the
// workspace has no list yet (so the next web app starts from it). Nothing else
// changes.
//
// Parameters: src is the text of smith.toml; project the app; s its settings.
func EnvironmentsToml(src, project string, s webenv.Settings) string {
	n := s.Normalized()
	table := EnvironmentsTable(project)
	src = configedit.SetKey(src, table, "environments", QuoteList(n.Names))
	src = configedit.SetKey(src, table, "default_environment", configedit.Quote(n.Default))
	src = configedit.SetKey(src, table, "output", configedit.Quote(n.Output))
	src = configedit.SetKey(src, table, "env_files", configedit.Quote(n.EnvFiles))
	if _, set := configedit.GetKey(src, "workspace", "environments"); !set {
		src = configedit.SetKey(src, "workspace", "environments", QuoteList(n.Names))
	}
	return src
}

// QuoteList renders a TOML array of strings.
func QuoteList(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = configedit.Quote(n)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// recordEnvironments writes the environments of a new web app into smith.toml
// and reloads the session, so the app's tasks are discovered with them.
func recordEnvironments(s *session.Session, project string, settings webenv.Settings) error {
	path := filepath.Join(s.Root, workspace.FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := EnvironmentsToml(string(raw), project, settings)
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err
	}
	return s.Reload()
}
