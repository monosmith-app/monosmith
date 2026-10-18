package publishing

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Docker runs docker; tests replace it. stdin, when set, is the command's
// input (the registry password for docker login).
var Docker = func(dir string, argv []string, stdin string) error {
	if stdin == "" {
		return toolrun.Run(dir, argv)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout, cmd.Stderr = toolrun.Log, toolrun.Log
	return cmd.Run()
}

// Inspect runs a docker command for its exit status only, so a manifest does
// not fill the release log; tests replace it.
var Inspect = func(dir string, argv []string) error {
	_, err := toolrun.Output(dir, argv)
	return err
}

// Publish builds a released app's image once, tagged with its version and
// latest under each registry [conventions.docker] names, and pushes them.
// With no registry, or push = false, the image is not pushed and the release
// goes on.
//
// It logs in to each registry first: ghcr.io with GITHUB_TOKEN (as
// GITHUB_ACTOR), Docker Hub with DOCKERHUB_USERNAME and DOCKERHUB_TOKEN,
// another registry with CONTAINER_USERNAME and CONTAINER_PASSWORD or the
// variables its entry names; without them the daemon's own login is used.
//
// A registry that already has the version's tag is left alone, so a re-run
// after one push failed pushes only to the registries that lack it. Every
// registry is tried; the error names each one that failed. Per registry the
// latest tag is pushed first and the version last, so the version tag shows
// the registry is complete.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if !images.Ships(workspaceRoot, params.Project) {
		return protocol.PublishResult{Skipped: "no Dockerfile"}, nil
	}
	t, err := images.TargetOf(params.Project, os.Getenv)
	if err != nil {
		return protocol.PublishResult{}, err
	}
	result := protocol.PublishResult{Registry: t.Where()}
	switch {
	case len(t.Registries) == 0:
		result.Skipped = "image built, not pushed: set [conventions.docker] registry in smith.toml to push it on release"
		return result, nil
	case !t.Push:
		result.Skipped = "image not pushed: push = false in smith.toml"
		return result, nil
	}
	if t.Legacy {
		_, _ = fmt.Fprintf(toolrun.Log, "docker: %s is deprecated; set [conventions.docker] registry = %q in smith.toml\n", images.LegacyRegistryEnv, t.Registries[0].URL)
	}
	if params.DryRun {
		var refs []string
		for _, r := range t.Registries {
			image := t.Reference(r)
			refs = append(refs, image+":"+params.Version, image+":latest")
		}
		result.Skipped = "would push " + strings.Join(refs, ", ")
		return result, nil
	}

	var pending []images.Registry
	var failed []string
	var had []string
	for _, r := range t.Registries {
		if err := login(workspaceRoot, r); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", r.URL, err))
			continue
		}
		if has(workspaceRoot, t.Reference(r)+":"+params.Version, r) {
			had = append(had, r.URL)
			continue
		}
		pending = append(pending, r)
	}
	if len(pending) > 0 {
		var tags []string
		for _, r := range pending {
			image := t.Reference(r)
			tags = append(tags, image+":"+params.Version, image+":latest")
		}
		if err := Docker(workspaceRoot, images.BuildArgs(params.Project, tags...), ""); err != nil {
			return result, fmt.Errorf("building the image: %w", err)
		}
	}
	var pushed []string
	for _, r := range pending {
		image := t.Reference(r)
		if err := push(workspaceRoot, image+":latest", image+":"+params.Version); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", r.URL, err))
			continue
		}
		pushed = append(pushed, r.URL)
	}
	if len(failed) > 0 {
		done := append(append([]string{}, had...), pushed...)
		hint := ""
		if len(done) > 0 {
			hint = "; " + strings.Join(done, ", ") + " already have " + params.Version + ", so a re-run pushes only to the others"
		}
		return result, errors.New("pushing the image failed for " + strings.Join(failed, "; ") + hint)
	}
	result.Registry = strings.Join(pushed, ", ")
	switch {
	case len(pushed) > 0:
		result.Published = true
		syncDescription(workspaceRoot, t)
	default:
		result.Skipped = "image " + params.Version + " already published to " + strings.Join(had, ", ")
	}
	return result, nil
}

// push pushes each tag, in order.
func push(workspaceRoot string, tags ...string) error {
	for _, tag := range tags {
		if err := Docker(workspaceRoot, []string{"docker", "push", tag}, ""); err != nil {
			return err
		}
	}
	return nil
}

// has reports whether a registry has the tag. Any failure to ask counts as
// "no": the push that follows then fails with its own, clearer error.
func has(workspaceRoot, tag string, r images.Registry) bool {
	argv := []string{"docker", "manifest", "inspect"}
	if images.Loopback(r.URL) {
		argv = append(argv, "--insecure")
	}
	return Inspect(workspaceRoot, append(argv, tag)) == nil
}

// login logs docker in to a registry with the credentials it takes, or
// leaves the daemon's login when none are set.
func login(workspaceRoot string, r images.Registry) error {
	user, password := credentialsOf(r)
	if user == "" {
		return nil
	}
	return Docker(workspaceRoot, []string{"docker", "login", r.Host(), "-u", user, "--password-stdin"}, password)
}

// credentialsOf reads a registry's user and password from the environment.
// GitHub's token logs in as the actor, or as x-access-token without one.
func credentialsOf(r images.Registry) (user, password string) {
	switch r.Login {
	case images.LoginNone:
		return "", ""
	case images.LoginGitHub:
		password = os.Getenv(r.PasswordEnv)
		if password == "" {
			return "", ""
		}
		if user = os.Getenv(r.UserEnv); user == "" {
			user = "x-access-token"
		}
		return user, password
	}
	return os.Getenv(r.UserEnv), os.Getenv(r.PasswordEnv)
}
