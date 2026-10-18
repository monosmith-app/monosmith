package recipereleases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// APIURL is GitHub's API; tests point it at a server.
var APIURL = "https://api.github.com"

// Latest asks GitHub for the newest published release of the recipe a
// location names, among the repository's releases tagged name@x.y.z. It
// returns the version, or an error when GitHub cannot be asked (offline, rate
// limited, no such repository) or the repository has no such release.
func Latest(client *http.Client, loc Location) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", APIURL, loc.Owner, loc.Repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var releases []struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	best := ""
	for _, r := range releases {
		m := tagPattern.FindStringSubmatch(r.Tag)
		if r.Draft || r.Prerelease || m == nil || m[1] != loc.Name {
			continue
		}
		if best == "" || Newer(m[2], best) {
			best = m[2]
		}
	}
	if best == "" {
		return "", fmt.Errorf("%s/%s has no release of %s", loc.Owner, loc.Repo, loc.Name)
	}
	return best, nil
}

// Newer reports whether version a is higher than b, both major.minor.patch.
func Newer(a, b string) bool {
	pa, pb := parts(a), parts(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func parts(v string) [3]int {
	var out [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		out[i], _ = strconv.Atoi(s)
	}
	return out
}
