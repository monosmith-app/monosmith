package ciannotations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIURL is GitHub's API; tests point it at a server.
var APIURL = "https://api.github.com"

// notStartedMarker is the start of the annotation GitHub writes when it does
// not start a job because of billing.
const notStartedMarker = "The job was not started because"

// maxCheckRuns caps the check runs read for one commit.
const maxCheckRuns = 30

// NotStarted reads the latest run of the default branch of owner/repo and
// returns the annotation that says a job was not started (a billing problem),
// or "" when there is none. It only reads. The error is set when GitHub
// cannot be asked (offline, bad token, no such repository); callers treat it
// as "nothing to report".
func NotStarted(client *http.Client, owner, repo, token string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	base := fmt.Sprintf("%s/repos/%s/%s", APIURL, url.PathEscape(owner), url.PathEscape(repo))
	var info struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := get(client, token, base, &info); err != nil {
		return "", err
	}
	var runs struct {
		Runs []struct {
			HeadSHA string `json:"head_sha"`
		} `json:"workflow_runs"`
	}
	if err := get(client, token, base+"/actions/runs?per_page=1&branch="+url.QueryEscape(info.DefaultBranch), &runs); err != nil {
		return "", err
	}
	if len(runs.Runs) == 0 {
		return "", nil
	}
	var checks struct {
		CheckRuns []struct {
			ID         int64  `json:"id"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := get(client, token, fmt.Sprintf("%s/commits/%s/check-runs?per_page=%d", base, url.PathEscape(runs.Runs[0].HeadSHA), maxCheckRuns), &checks); err != nil {
		return "", err
	}
	for _, c := range checks.CheckRuns {
		if c.Conclusion != "failure" {
			continue
		}
		var notes []struct {
			Message string `json:"message"`
		}
		if err := get(client, token, fmt.Sprintf("%s/check-runs/%d/annotations", base, c.ID), &notes); err != nil {
			return "", err
		}
		for _, n := range notes {
			if strings.HasPrefix(strings.TrimSpace(n.Message), notStartedMarker) {
				return strings.TrimSpace(n.Message), nil
			}
		}
	}
	return "", nil
}

func get(client *http.Client, token, rawURL string, into any) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "smith")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub answered %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(into)
}
