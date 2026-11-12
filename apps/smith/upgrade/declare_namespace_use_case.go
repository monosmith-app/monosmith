package upgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

// declareNamespace keeps the names of new projects in a workspace made before
// [workspace] namespace existed. Such a workspace gave every language but npm
// the word of its scope (`scope = "@acme"` made acme-core, acme_core and
// Acme.Core); the namespace now defaults to the workspace name, so without this
// step the next project would be named after the workspace instead. When the
// scope and the name differ, it adds namespace = "<scope without @>" and
// leaves the scope as it is. The edit is in the plan smith shows before it
// writes. With check it changes nothing.
func declareNamespace(s *session.Session, check bool, r *Result) error {
	ws := s.Config.Workspace
	namespace := strings.TrimPrefix(strings.TrimSpace(ws.Scope), "@")
	if ws.Namespace != "" || namespace == "" || projectnaming.Kebab(namespace) == projectnaming.Kebab(ws.Name) {
		return nil
	}
	r.Migrations = append(r.Migrations, words.Say("naming.upgrade.kept", "namespace", namespace, "name", ws.Name, "fromName", projectnaming.Kebab(ws.Name), "fromNamespace", projectnaming.Kebab(namespace)))
	addFile(r, workspace.FileName)
	if check {
		return nil
	}
	file := filepath.Join(s.Root, workspace.FileName)
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(tomltext.SetString(string(raw), "workspace", "namespace", namespace)), 0o644)
}
