package releasepreview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpClient sends the requests; tests point it at a local server.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// Upsert creates smith's comment on the pull request, or replaces the one an
// earlier run left (found by its Marker).
func Upsert(pr PullRequest, body string) (string, error) {
	switch pr.Provider {
	case "github":
		return upsertGitHub(pr, body)
	case "azure":
		return upsertAzure(pr, body)
	case "gitlab":
		return upsertGitLab(pr, body)
	case "bitbucket":
		return upsertBitbucket(pr, body)
	}
	return "", fmt.Errorf("unknown provider %q", pr.Provider)
}

func upsertGitHub(pr PullRequest, body string) (string, error) {
	auth := "Bearer " + pr.Token
	base := fmt.Sprintf("%s/repos/%s/issues", pr.API, pr.Repo)
	for page := 1; page <= 10; page++ {
		var comments []struct {
			ID   int64  `json:"id"`
			Body string `json:"body"`
		}
		if err := call(http.MethodGet, fmt.Sprintf("%s/%d/comments?per_page=100&page=%d", base, pr.Number, page), auth, nil, &comments); err != nil {
			return "", err
		}
		for _, c := range comments {
			if strings.Contains(c.Body, Marker) {
				return "updated", call(http.MethodPatch, fmt.Sprintf("%s/comments/%d", base, c.ID), auth, map[string]string{"body": body}, nil)
			}
		}
		if len(comments) < 100 {
			break
		}
	}
	return "created", call(http.MethodPost, fmt.Sprintf("%s/%d/comments", base, pr.Number), auth, map[string]string{"body": body}, nil)
}

func upsertAzure(pr PullRequest, body string) (string, error) {
	auth := "Bearer " + pr.Token
	threads := fmt.Sprintf("%s%s/_apis/git/repositories/%s/pullRequests/%d/threads", pr.Collection, pr.Project, pr.RepoID, pr.Number)
	var list struct {
		Value []struct {
			ID       int64 `json:"id"`
			Comments []struct {
				ID      int64  `json:"id"`
				Content string `json:"content"`
			} `json:"comments"`
		} `json:"value"`
	}
	if err := call(http.MethodGet, threads+"?api-version=7.1", auth, nil, &list); err != nil {
		return "", err
	}
	for _, t := range list.Value {
		if len(t.Comments) > 0 && strings.Contains(t.Comments[0].Content, Marker) {
			url := fmt.Sprintf("%s/%d/comments/%d?api-version=7.1", threads, t.ID, t.Comments[0].ID)
			return "updated", call(http.MethodPatch, url, auth, map[string]string{"content": body}, nil)
		}
	}
	// Closed (status 4): a preview never waits to be resolved.
	thread := map[string]any{
		"comments": []map[string]any{{"parentCommentId": 0, "content": body, "commentType": 1}},
		"status":   4,
	}
	return "created", call(http.MethodPost, threads+"?api-version=7.1", auth, thread, nil)
}

func upsertGitLab(pr PullRequest, body string) (string, error) {
	auth := "Bearer " + pr.Token
	notes := fmt.Sprintf("%s/projects/%s/merge_requests/%d/notes", pr.API, url.PathEscape(pr.ProjectID), pr.Number)
	for page := 1; page <= 10; page++ {
		var list []struct {
			ID   int64  `json:"id"`
			Body string `json:"body"`
		}
		if err := call(http.MethodGet, fmt.Sprintf("%s?per_page=100&page=%d", notes, page), auth, nil, &list); err != nil {
			return "", err
		}
		for _, n := range list {
			if strings.Contains(n.Body, Marker) {
				return "updated", call(http.MethodPut, fmt.Sprintf("%s/%d", notes, n.ID), auth, map[string]string{"body": body}, nil)
			}
		}
		if len(list) < 100 {
			break
		}
	}
	return "created", call(http.MethodPost, notes, auth, map[string]string{"body": body}, nil)
}

func upsertBitbucket(pr PullRequest, body string) (string, error) {
	auth := "Bearer " + pr.Token
	comments := fmt.Sprintf("%s/repositories/%s/pullrequests/%d/comments", pr.API, pr.Repo, pr.Number)
	content := map[string]any{"content": map[string]string{"raw": body}}
	next := comments + "?pagelen=100"
	for page := 1; next != "" && page <= 10; page++ {
		var list struct {
			Values []struct {
				ID      int64 `json:"id"`
				Content struct {
					Raw string `json:"raw"`
				} `json:"content"`
			} `json:"values"`
			Next string `json:"next"`
		}
		if err := call(http.MethodGet, next, auth, nil, &list); err != nil {
			return "", err
		}
		for _, c := range list.Values {
			if strings.Contains(c.Content.Raw, Marker) {
				return "updated", call(http.MethodPut, fmt.Sprintf("%s/%d", comments, c.ID), auth, content, nil)
			}
		}
		next = list.Next
	}
	return "created", call(http.MethodPost, comments, auth, content, nil)
}

func call(method, url, auth string, in, out any) error {
	var payload io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s %s: %s %s", method, strings.Split(url, "?")[0], resp.Status, strings.TrimSpace(string(detail)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
