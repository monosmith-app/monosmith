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
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// hubURL is Docker Hub's API; tests point it at a local server.
var hubURL = "https://hub.docker.com"

// hubClient talks to Docker Hub's API.
var hubClient = &http.Client{Timeout: time.Minute}

// hubDescriptionLimit is the longest repository description Docker Hub takes.
const hubDescriptionLimit = 25000

// syncDescription sets the repository description of the image on Docker Hub
// from the file description_from names, for each docker.io registry. The
// image is already pushed, so a failure here (a token that cannot edit the
// repository, a file that is gone) is a warning and never fails the release.
// Registries other than Docker Hub are left alone.
func syncDescription(workspaceRoot string, t images.Target) {
	if t.DescriptionFrom == "" {
		return
	}
	var hub []images.Registry
	for _, r := range t.Registries {
		if images.LoginFor(r.URL) == images.LoginDockerHub {
			hub = append(hub, r)
		}
	}
	if len(hub) == 0 {
		return
	}
	text, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(t.DescriptionFrom)))
	if err != nil {
		warn("not setting the Docker Hub description: %v", err)
		return
	}
	description := string(text)
	if len(description) > hubDescriptionLimit {
		warn("%s is longer than the %d characters Docker Hub takes; the description is cut", t.DescriptionFrom, hubDescriptionLimit)
		description = strings.ToValidUTF8(description[:hubDescriptionLimit], "")
	}
	for _, r := range hub {
		if err := setHubDescription(r, t.Image, description); err != nil {
			warn("the Docker Hub description of %s/%s was not set: %v. Editing a description may need an access token with Read, Write & Delete permission", r.URL, t.Image, err)
		}
	}
}

func warn(format string, args ...any) {
	_, _ = fmt.Fprintf(toolrun.Log, "docker: warning: "+format+"\n", args...)
}

// setHubDescription logs in to Docker Hub with the registry's credentials and
// sets the repository's full description.
func setHubDescription(r images.Registry, image, description string) error {
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
	return hubRequest(http.MethodPatch, "/v2/repositories/"+namespace+"/"+image+"/", login.Token, map[string]string{"full_description": description}, nil)
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
