package publishing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// hubURL is Docker Hub's API; tests point it at a local server.
var hubURL = "https://hub.docker.com"

// hubClient talks to Docker Hub's API.
var hubClient = &http.Client{Timeout: time.Minute}

// hubDescriptionLimit is the longest repository description Docker Hub takes.
const hubDescriptionLimit = 25000

// DescriptionResult says what setting a Docker Hub description did.
type DescriptionResult struct {
	// Set are the repositories whose description was set, or would be.
	Set []string
	// Failed are the repositories it could not be set for, with the reason.
	Failed []string
	// Skipped says why nothing was attempted.
	Skipped string
	// Short is the short description sent, or that would be, fitted to Docker
	// Hub's limit; empty when none is set.
	Short string
	// Full names the file the full description comes from; empty when none is
	// set.
	Full string
}

// SyncDescription sets the Docker Hub descriptions of an app's image, for
// each docker.io registry: the full one from the file description_from names,
// the short one from short_description or the project's description. It runs
// without a
// release: the call is an idempotent PATCH, so running it again is harmless.
// A failure is in the result and logged as a warning, never an error; the
// error is for settings that contradict each other. With dryRun it only
// names the repositories.
func SyncDescription(workspaceRoot string, p protocol.Project, dryRun bool) (DescriptionResult, error) {
	if !images.Ships(workspaceRoot, p) {
		return DescriptionResult{Skipped: "no Dockerfile"}, nil
	}
	t, err := images.TargetOf(p, os.Getenv)
	if err != nil {
		return DescriptionResult{}, err
	}
	return describe(workspaceRoot, t, dryRun), nil
}

// syncDescription sets the repository description of the image on Docker Hub
// from the file description_from names, for each docker.io registry. The
// image is already pushed, so a failure here (a token that cannot edit the
// repository, a file that is gone) is a warning and never fails the release.
// Registries other than Docker Hub are left alone.
func syncDescription(workspaceRoot string, t images.Target) {
	describe(workspaceRoot, t, false)
}

func describe(workspaceRoot string, t images.Target, dryRun bool) DescriptionResult {
	var res DescriptionResult
	if t.DescriptionFrom == "" && strings.TrimSpace(t.ShortDescription) == "" {
		res.Skipped = "no description_from or short description"
		return res
	}
	var hub []images.Registry
	for _, r := range t.Registries {
		if images.LoginFor(r.URL) == images.LoginDockerHub {
			hub = append(hub, r)
		}
	}
	if len(hub) == 0 {
		res.Skipped = "no docker.io registry"
		return res
	}
	// The fields to send: only the ones set, so a missing full description
	// leaves what is on Docker Hub alone, and so does a missing short one.
	fields := map[string]string{}
	if t.DescriptionFrom != "" {
		text, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(t.DescriptionFrom)))
		if err != nil {
			warn("not setting the Docker Hub description: %v", err)
			res.Failed = append(res.Failed, t.DescriptionFrom+": "+err.Error())
		} else {
			description := string(text)
			if len(description) > hubDescriptionLimit {
				warn("%s is longer than the %d characters Docker Hub takes; the description is cut", t.DescriptionFrom, hubDescriptionLimit)
				description = strings.ToValidUTF8(description[:hubDescriptionLimit], "")
			}
			fields["full_description"] = description
			res.Full = t.DescriptionFrom
		}
	}
	if strings.TrimSpace(t.ShortDescription) != "" {
		short, cut := images.FitShortDescription(t.ShortDescription)
		if cut {
			warn("the short description is longer than the %d characters Docker Hub takes; it is cut to %q", images.ShortDescriptionLimit, short)
		}
		fields["description"] = short
		res.Short = short
	}
	if len(fields) == 0 {
		return res
	}
	for _, r := range hub {
		repository := r.URL + "/" + t.Image
		if dryRun {
			res.Set = append(res.Set, repository)
			continue
		}
		if err := setHubDescription(r, t.Image, fields); err != nil {
			warn("the Docker Hub description of %s/%s was not set: %v. Editing a description may need an access token with Read, Write & Delete permission", r.URL, t.Image, err)
			res.Failed = append(res.Failed, repository+": "+err.Error())
			continue
		}
		res.Set = append(res.Set, repository)
	}
	return res
}

func warn(format string, args ...any) {
	_, _ = fmt.Fprintf(toolrun.Log, "docker: warning: "+format+"\n", args...)
}

// setHubDescription logs in to Docker Hub with the registry's credentials and
// sets the repository's descriptions in one call: fields holds
// full_description and description (the short one), whichever are set.
func setHubDescription(r images.Registry, image string, fields map[string]string) error {
	user, password := credentialsOf(r)
	if user == "" || password == "" {
		return fmt.Errorf("%s and %s are not both set", r.UserEnv, r.PasswordEnv)
	}
	namespace := strings.TrimPrefix(strings.TrimPrefix(r.URL, r.Host()), "/")
	if namespace == "" {
		return fmt.Errorf("%s names no namespace", r.URL)
	}
	var login struct {
		Token string `json:"token"`
	}
	if err := hubRequest(http.MethodPost, "/v2/users/login/", "", map[string]string{"username": user, "password": password}, &login); err != nil {
		return fmt.Errorf("signing in: %w", err)
	}
	if login.Token == "" {
		return fmt.Errorf("signing in: no token in the answer")
	}
	return hubRequest(http.MethodPatch, "/v2/repositories/"+namespace+"/"+image+"/", login.Token, fields, nil)
}

// hubRequest sends a JSON request to Docker Hub and decodes the answer into
// out when it is not nil.
func hubRequest(method, path, token string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, hubURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "JWT "+token)
	}
	resp, err := hubClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	answer, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(answer)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(answer, out)
}
