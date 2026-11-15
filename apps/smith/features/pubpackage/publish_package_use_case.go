package pubpackage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Result says what happened to the package.
type Result struct {
	// Published or Existing is name@version.
	Published, Existing string
	// Skipped are the archives Dart does not run on.
	Skipped []string
}

// dart runs the Dart CLI; tests replace it.
var dart = func(dir string, args ...string) (stdout, stderr string, err error) {
	return toolrun.Capture(dir, append([]string{"dart"}, args...))
}

// published lists the versions pub.dev has of a package; tests replace it.
var published = func(name string) ([]string, error) {
	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Get("https://pub.dev/api/packages/" + name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pub.dev: %s", resp.Status)
	}
	var body struct {
		Versions []struct{ Version string } `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var versions []string
	for _, v := range body.Versions {
		versions = append(versions, v.Version)
	}
	return versions, nil
}

// Has reports whether pub.dev has a version of a package.
func Has(name, version string) (bool, error) {
	versions, err := published(name)
	if err != nil {
		return false, err
	}
	return slices.Contains(versions, version), nil
}

// Publish builds the package and publishes it with `dart pub publish`,
// unless pub.dev has its version, so a re-run completes a release. It
// publishes with whatever credentials dart has: a login (`dart pub publish`
// asks for one the first time) or a CI token. A dry run only names it.
func Publish(root string, r Release, dryRun bool) (Result, error) {
	var res Result
	if r.Name == "" {
		return res, nil
	}
	id := r.Name + "@" + r.Version
	if dryRun {
		res.Published = id
		return res, nil
	}
	has, err := Has(r.Name, r.Version)
	if err != nil {
		return res, err
	}
	if has {
		res.Existing = id
		return res, nil
	}
	dir, skipped, err := Build(root, r)
	res.Skipped = skipped
	if err != nil {
		return res, err
	}
	// dart pub publish leaves out what git ignores, and the package is built
	// under dist/, which a workspace ignores: it is published from a copy
	// outside the repository.
	staged, err := os.MkdirTemp("", "smith-pub-")
	if err != nil {
		return res, err
	}
	defer func() { _ = os.RemoveAll(staged) }()
	if err := os.CopyFS(staged, os.DirFS(dir)); err != nil {
		return res, err
	}
	if stdout, stderr, err := dart(staged, "pub", "publish", "--force"); err != nil {
		// dart prints the validation errors on stdout.
		return res, fmt.Errorf("dart pub publish %s: %w\n%s", id, err, strings.TrimSpace(stdout+"\n"+stderr))
	}
	res.Published = id
	return res, nil
}
