package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// secretWord finds a name that looks like a secret.
var secretWord = regexp.MustCompile(`(?i)KEY|SECRET|TOKEN`)

// viteName reads the name of a VITE_ line of an env file.
var viteName = regexp.MustCompile(`^\s*(?:export\s+)?(VITE_[A-Za-z0-9_]*)\s*=`)

// webEnvironmentFindings checks the web apps that are built once per
// environment (ADR 0081): each environment has its .env.<env> file (an app
// that commits only .env.example needs that file instead), and no VITE_ name
// looks like a secret, because every VITE_ value is written into the browser
// bundle. It returns nothing when no project has environments.
//
// Parameters: root is the workspace folder; projects are the discovered ones.
func webEnvironmentFindings(root string, projects []*projectgraph.Project) []protocol.Finding {
	var missing, secrets []string
	var fixFiles []protocol.FileEdit
	var apps []string
	for _, p := range projects {
		settings, ok := webenv.FromMap(p.Settings[webenv.PluginName])
		if !ok {
			continue
		}
		apps = append(apps, p.Name)
		dir := filepath.Join(root, filepath.FromSlash(p.Root))
		if settings.EnvFiles == webenv.FilesExample {
			if _, err := os.Stat(filepath.Join(dir, ".env.example")); err != nil {
				missing = append(missing, p.Root+"/.env.example")
				fixFiles = append(fixFiles, protocol.FileEdit{Path: p.Root + "/.env.example", Content: webenv.Files(settings)[".env.example"]})
			}
		}
		for _, env := range settings.Names {
			file := webenv.EnvFileName(env)
			raw, err := os.ReadFile(filepath.Join(dir, file))
			if err != nil {
				if settings.EnvFiles == webenv.FilesCommit {
					missing = append(missing, p.Root+"/"+file)
					fixFiles = append(fixFiles, protocol.FileEdit{Path: p.Root + "/" + file, Content: webenv.EnvFile(env, settings)})
				}
				continue
			}
			for _, name := range SecretLookingNames(string(raw)) {
				secrets = append(secrets, fmt.Sprintf("%s in %s/%s", name, p.Root, file))
			}
		}
	}
	if len(apps) == 0 {
		return nil
	}
	sort.Strings(missing)
	sort.Strings(secrets)
	have := protocol.Finding{
		Check: "every environment of a web app has its env file", OK: len(missing) == 0, Warning: true, Area: AreaConfig,
		Problem: problemWhen(len(missing) > 0, "env files are missing: "+leadProblem(missing)),
		Detail:  "Vite reads .env.<environment> when it builds that environment. Missing: " + strings.Join(missing, ", "),
		Remedy:  "make the files (smith doctor --fix writes a starter for each), or remove the environment (smith environments remove <project> <name>)",
	}
	if len(fixFiles) > 0 {
		have.Fix = &protocol.Fix{Description: "Write the missing env files of the web apps", Files: fixFiles}
	}
	return []protocol.Finding{have, {
		Check: "no VITE_ value of a web app looks like a secret", OK: len(secrets) == 0, Warning: true, Area: AreaConfig,
		Problem: problemWhen(len(secrets) > 0, "VITE_ names that look like secrets: "+leadProblem(secrets)),
		Detail:  "every VITE_ value is written into the browser bundle, where anyone can read it. Names with KEY, SECRET or TOKEN in them: " + strings.Join(secrets, ", "),
		Remedy:  "keep the secret on a server and let the app ask the server, or rename the variable if it is not a secret (a public key is fine, a private key is not); a secret the build needs goes through [ci] secrets in smith.toml and never into a VITE_ name",
	}}
}

// SecretLookingNames lists the VITE_ names of an env file that have KEY,
// SECRET or TOKEN in them, so a person looks at them before they ship.
func SecretLookingNames(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if m := viteName.FindStringSubmatch(strings.TrimSpace(line)); m != nil && secretWord.MatchString(strings.TrimPrefix(m[1], "VITE_")) {
			out = append(out, m[1])
		}
	}
	return out
}
