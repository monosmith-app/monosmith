package adoption

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/manifestoffer"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// writeRootManifest has the language plugin write the manifest of a root of
// sources (language), named after and with the kind of the destination folder (apps/ is an
// app, libs/ a lib, packages/ a package; elsewhere what the sources look like,
// said in the output). git add stages the files, so git mv moves them and
// git reset --hard takes them back. The root is then identified again and the
// move into dir planned for the project.
//
// Parameters: r is the report; c is the candidate; language is the plugin's
// language; dir is the settled folder; version is smith's; out gets the progress.
// Returns: the identified candidate and its move plan, or an error from the
// plugin, from git, or from the plan.
func writeRootManifest(r Report, c adoptmove.Candidate, language, dir, version string, out io.Writer) (adoptmove.Candidate, adoptmove.Plan, error) {
	kind := projectkind.Resolve("", dir, "")
	if kind == "" {
		kind = protocol.ProjectLib
		if c.App {
			kind = protocol.ProjectApp
		}
		fmt.Fprintf(out, "  kind: %s, from what the sources look like (%s is outside apps/, libs/ and packages/); smith kind changes it\n", kind, dir)
	}
	// The project is named after the folder it goes into.
	c.Name = adoptmove.DeriveName("", path.Base(dir))
	cfg, err := previewConfig(r)
	if err != nil {
		return c, adoptmove.Plan{}, err
	}
	s, err := session.Open(session.Options{Dir: r.Root, Version: version, Log: io.Discard, Preview: &cfg})
	if err != nil {
		return c, adoptmove.Plan{}, err
	}
	w, err := manifestoffer.Write(s, language, kind, c.Name, ".")
	s.Close()
	if err != nil {
		return c, adoptmove.Plan{}, err
	}
	fmt.Fprintf(out, "  manifest: wrote %s from smith add's template; the sources keep their layout\n", strings.Join(w.Files, ", "))
	if _, err := gitclient.New(r.Root).Run(append([]string{"add", "--"}, w.Files...)...); err != nil {
		return c, adoptmove.Plan{}, fmt.Errorf("git add of the new manifest failed: %w", err)
	}
	found, err := adoptmove.Detect(r.Root, r.Claims)
	if err != nil {
		return c, adoptmove.Plan{}, fmt.Errorf("the manifest was written, but smith does not see the project at the root (%w): git reset --hard takes it back", err)
	}
	found.Name = c.Name
	plan, err := PlanFor(r.Root, found, dir)
	return found, plan, err
}
