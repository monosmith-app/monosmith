package pluginmarket

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/plugininstall"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// maxIndexBytes caps the index: a list of plugins is text.
const maxIndexBytes = 8 << 20

// githubHosts are the only hosts that ever receive a token.
var githubHosts = map[string]bool{
	"raw.githubusercontent.com": true,
	"api.github.com":            true,
	"github.com":                true,
}

// FetchIndex reads the index at location (an https URL, an http URL on this
// machine, or a file path) and checks it. A GitHub host that answers 401, 403
// or 404 is asked again with a token, which is the one smith finds with
// GitHubToken; that is how the index of a private repository is read. A token
// goes to GitHub hosts only, never to a location of another host, and a public
// index is read without one.
func FetchIndex(ctx context.Context, client *http.Client, location string, token func() string) (Index, error) {
	data, err := fetchBytes(ctx, client, location, token)
	if err != nil {
		return Index{}, err
	}
	idx, err := ParseIndex(data)
	if err != nil {
		return Index{}, problem.New(err.Error(), "check the marketplace location: "+location)
	}
	return idx, nil
}

func fetchBytes(ctx context.Context, client *http.Client, location string, token func() string) ([]byte, error) {
	if !strings.HasPrefix(location, "https://") && !strings.HasPrefix(location, "http://") {
		if strings.Contains(location, "://") {
			return nil, problem.New(location+" is not a marketplace location", "use an https URL or a file path")
		}
		data, err := os.ReadFile(location)
		if err != nil {
			return nil, problem.New("cannot read the marketplace index: "+err.Error(), "check "+LocationEnv+" and [plugins] marketplace in smith.toml")
		}
		return data, nil
	}
	if err := plugininstall.CheckURL(location); err != nil {
		return nil, problem.New(err.Error())
	}
	u, _ := url.Parse(location)
	onGitHub := githubHosts[u.Hostname()]

	data, status, err := get(ctx, client, location, "")
	if err != nil {
		return nil, problem.New(fmt.Sprintf("cannot reach the marketplace at %s: %v", location, err), "check your network, or set "+LocationEnv)
	}
	if status == http.StatusOK {
		return data, nil
	}
	if onGitHub && (status == http.StatusNotFound || status == http.StatusUnauthorized || status == http.StatusForbidden) {
		tok := ""
		if token != nil {
			tok = token()
		}
		if tok == "" {
			return nil, problem.New(
				fmt.Sprintf("the marketplace at %s answered %d: it is private, or it does not exist", location, status),
				"set GITHUB_TOKEN (or GH_TOKEN) to a token that can read it",
				"or sign in with `gh auth login`",
				"or point "+LocationEnv+" at another index")
		}
		data, status, err = get(ctx, client, location, tok)
		if err != nil {
			return nil, problem.New(fmt.Sprintf("cannot reach the marketplace at %s: %v", location, err))
		}
		if status == http.StatusOK {
			return data, nil
		}
		return nil, problem.New(fmt.Sprintf("the marketplace at %s answered %d even with your token", location, status),
			"check that the token can read the repository")
	}
	return nil, problem.New(fmt.Sprintf("the marketplace at %s answered %d", location, status), "check "+LocationEnv+" and [plugins] marketplace in smith.toml")
}

// get returns the body and status of one request, sending token when given.
func get(ctx context.Context, client *http.Client, rawURL, token string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "smith")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxIndexBytes+1))
	if err != nil {
		return nil, 0, err
	}
	if len(data) > maxIndexBytes {
		return nil, 0, fmt.Errorf("the index is larger than %d bytes", maxIndexBytes)
	}
	return data, resp.StatusCode, nil
}

// GitHubToken finds a token for a private marketplace: GITHUB_TOKEN, then
// GH_TOKEN, then the output of `gh auth token` when the GitHub CLI is
// installed and signed in. It returns "" when there is none.
func GitHubToken() string {
	for _, name := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	path, err := exec.LookPath("gh")
	if err != nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
