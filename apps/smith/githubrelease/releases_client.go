package githubrelease

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrNoToken is returned when no GitHub token is available.
var ErrNoToken = errors.New("GITHUB_TOKEN (or GH_TOKEN) is not set; GitHub Releases need a token with contents: write")

// Client talks to the GitHub REST API.
type Client struct {
	Token string
	// API is the REST root, e.g. https://api.github.com.
	API  string
	HTTP *http.Client
}

// Release is the part of a GitHub Release smith uses.
type Release struct {
	ID        int64   `json:"id"`
	HTMLURL   string  `json:"html_url"`
	UploadURL string  `json:"upload_url"`
	Assets    []Asset `json:"assets"`
}

// Asset is a file attached to a release.
type Asset struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FromEnvironment builds a client from GITHUB_TOKEN or GH_TOKEN and
// GITHUB_API_URL (set on GitHub Enterprise runners).
func FromEnvironment() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		return nil, ErrNoToken
	}
	api := os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}
	return &Client{Token: token, API: strings.TrimSuffix(api, "/"), HTTP: &http.Client{Timeout: 10 * time.Minute}}, nil
}

// ReleaseByTag returns the release of a tag; found is false when none exists.
func (c *Client) ReleaseByTag(repo Repository, tag string) (Release, bool, error) {
	var r Release
	status, err := c.do(http.MethodGet, c.API+"/repos/"+repo.String()+"/releases/tags/"+url.PathEscape(tag), "application/json", nil, &r)
	if status == http.StatusNotFound {
		return Release{}, false, nil
	}
	return r, err == nil, err
}

// CreateRelease publishes a release for an existing tag.
func (c *Client) CreateRelease(repo Repository, tag, name, notes string, prerelease bool) (Release, error) {
	body, _ := json.Marshal(map[string]any{"tag_name": tag, "name": name, "body": notes, "prerelease": prerelease})
	var r Release
	_, err := c.do(http.MethodPost, c.API+"/repos/"+repo.String()+"/releases", "application/json", body, &r)
	return r, err
}

// UploadAsset attaches a file to a release under its base name.
func (c *Client) UploadAsset(r Release, file string) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	target, _, _ := strings.Cut(r.UploadURL, "{")
	target += "?name=" + url.QueryEscape(filepath.Base(file))
	contentType := "application/octet-stream"
	if strings.HasSuffix(file, ".txt") {
		contentType = "text/plain"
	}
	_, err = c.do(http.MethodPost, target, contentType, raw, nil)
	return err
}

// DownloadAsset writes a release asset's content to a file.
func (c *Client) DownloadAsset(repo Repository, a Asset, to string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/releases/assets/%d", c.API, repo, a.ID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/octet-stream")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("downloading %s: %s: %s", a.Name, resp.Status, strings.TrimSpace(string(raw)))
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// DispatchWorkflow starts a workflow_dispatch run of a workflow file on a ref
// (a branch or tag), with its inputs.
func (c *Client) DispatchWorkflow(repo Repository, workflow, ref string, inputs map[string]string) error {
	body, _ := json.Marshal(map[string]any{"ref": ref, "inputs": inputs})
	_, err := c.do(http.MethodPost, c.API+"/repos/"+repo.String()+"/actions/workflows/"+url.PathEscape(workflow)+"/dispatches", "application/json", body, nil)
	return err
}

func (c *Client) do(method, target, contentType string, body []byte, out any) (int, error) {
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("%s %s: %s: %s", method, target, resp.Status, strings.TrimSpace(string(raw)))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return resp.StatusCode, fmt.Errorf("%s %s: %w", method, target, err)
		}
	}
	return resp.StatusCode, nil
}
