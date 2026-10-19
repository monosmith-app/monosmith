package releasing

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// MavenRunner runs mvn; tests replace it.
var MavenRunner = func(dir string, argv []string, env ...string) error { return toolrun.Run(dir, argv, env...) }

// skipped are directories no workspace pom lives in.
var skipped = map[string]bool{".git": true, protocol.BuildDir: true, "node_modules": true, ".smith": true}

// WriteVersion records a release's version in the module's pom and in every
// workspace pom that depends on the module with a version, or in a Gradle
// module's script.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	if script := gradleScript(workspaceRoot, params.Project.Root); script != "" {
		return writeGradleVersion(workspaceRoot, script, params)
	}
	result := protocol.VersionResult{Files: []string{}}
	own := path.Join(params.Project.Root, "pom.xml")
	group, artifact, _ := strings.Cut(coordinates(params.Project), ":")
	err := filepath.WalkDir(workspaceRoot, func(full string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipped[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "pom.xml" {
			return nil
		}
		rel, _ := filepath.Rel(workspaceRoot, full)
		rel = filepath.ToSlash(rel)
		src, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		out, changed := string(src), false
		if rel == own {
			out, changed = SetVersion(out, params.Version)
		}
		if artifact != "" {
			var dep bool
			out, dep = SetDependencyVersion(out, group, artifact, params.Version)
			changed = changed || dep
		}
		if !changed {
			return nil
		}
		result.Files = append(result.Files, rel)
		if params.DryRun {
			return nil
		}
		return os.WriteFile(full, []byte(out), 0o644)
	})
	return result, err
}

// coordinates is groupId:artifactId among a module's exports (a Gradle
// module also exports its :project:path).
func coordinates(p protocol.Project) string {
	for _, e := range p.Exports {
		if !strings.HasPrefix(e, ":") {
			return e
		}
	}
	return ""
}

// MavenCentral is where versions are read when MAVEN_REPOSITORY_URL is not
// set; tests replace it.
var MavenCentral = "https://repo1.maven.org/maven2"

// settings is the Maven settings Publish writes: the repository's
// credentials, read from the environment by Maven itself.
const settings = `<settings>
  <servers>
    <server>
      <id>smith</id>
      <username>${env.MAVEN_USERNAME}</username>
      <password>${env.MAVEN_PASSWORD}</password>
    </server>
  </servers>
</settings>
`

// Publish deploys a package, with Maven or Gradle, to the repository
// smith.toml names for it ([registries.<name>] maven, signed in as its auth
// says: GitHub Packages with the job's token, a token, Azure Artifacts), else
// to MAVEN_REPOSITORY_URL with the credentials in MAVEN_USERNAME and
// MAVEN_PASSWORD. Apps and libraries release by their tag.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return protocol.PublishResult{Skipped: "JVM release assets are not built yet: the app is released by its tag"}, nil
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	target, named, err := registries.Resolve(params.Project, os.Getenv, func() string { return gitRemote(workspaceRoot) })
	repository := os.Getenv("MAVEN_REPOSITORY_URL")
	if named {
		repository = target.URL
	}
	result := protocol.PublishResult{Registry: repository}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	if err != nil {
		return result, err
	}
	var env []string
	serverID, useSettings := registries.ServerSmith, true
	if named {
		user, password, err := target.Credentials(os.Getenv)
		if err != nil {
			return result, err
		}
		env = target.Env(user, password)
		// Azure Artifacts: MavenAuthenticate@0 wrote the feed's server into
		// the user's settings, which smith's own would hide.
		serverID, useSettings = target.ServerID, target.Auth != protocol.AuthAzureArtifacts
	}
	if repository == "" {
		return result, fmt.Errorf("set MAVEN_REPOSITORY_URL to the repository %s deploys to", params.Project.Name)
	}
	if gradleScript(workspaceRoot, params.Project.Root) != "" {
		return publishGradle(workspaceRoot, params.Project, result, env)
	}
	dir := filepath.Join(workspaceRoot, protocol.BuildDir, "maven")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}
	selector := params.Project.Root
	if c := coordinates(params.Project); c != "" {
		selector = c
	}
	argv := []string{"mvn", "-B", "-q", "-pl", selector, "deploy", "-DskipTests"}
	if useSettings {
		settingsFile := filepath.Join(dir, "settings.xml")
		if err := os.WriteFile(settingsFile, []byte(settings), 0o600); err != nil {
			return result, err
		}
		argv = append(argv, "-s", settingsFile)
	}
	argv = append(argv, "-DaltDeploymentRepository="+serverID+"::"+repository)
	if err := MavenRunner(workspaceRoot, argv, env...); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// Published reads the newest release from the repository's
// maven-metadata.xml: the registry smith.toml names for the package (signed
// in with its credentials when it has them, as GitHub Packages needs even to
// read), else MAVEN_REPOSITORY_URL, else Maven Central.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	group, artifact, ok := strings.Cut(coordinates(params.Project), ":")
	if params.Project.Kind != protocol.ProjectPackage || !ok {
		return protocol.PublishedResult{}, nil
	}
	repository := strings.TrimSuffix(os.Getenv("MAVEN_REPOSITORY_URL"), "/")
	var user, password string
	target, named, resolveErr := registries.Resolve(params.Project, os.Getenv, func() string { return gitRemote("") })
	if named {
		repository = target.URL
		if u, p, err := target.Credentials(os.Getenv); err == nil {
			user, password = u, p
		}
	}
	if repository == "" {
		repository = MavenCentral
	}
	result := protocol.PublishedResult{Registry: repository}
	if resolveErr != nil {
		result.Unreachable, result.Detail = true, resolveErr.Error()
		return result, nil
	}
	url := repository + "/" + strings.ReplaceAll(group, ".", "/") + "/" + artifact + "/maven-metadata.xml"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	if password != "" {
		req.SetBasicAuth(user, password)
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	defer func() { _ = resp.Body.Close() }()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return result, nil
	case resp.StatusCode != http.StatusOK:
		result.Unreachable, result.Detail = true, resp.Status
		return result, nil
	}
	var metadata struct {
		Release string `xml:"versioning>release"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	result.Version = metadata.Release
	return result, nil
}

// gitRemote is the URL of the workspace's origin remote, or empty.
func gitRemote(workspaceRoot string) string {
	out, err := toolrun.Output(workspaceRoot, []string{"git", "remote", "get-url", "origin"})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
