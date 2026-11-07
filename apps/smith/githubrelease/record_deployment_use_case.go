package githubrelease

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Deployment is one entry of the repository's Deployments sidebar: a registry
// a release published a package to.
type Deployment struct {
	// Environment is the registry's name (npm, PyPI, NuGet, RubyGems).
	Environment string
	// Ref is the release tag the package was published from.
	Ref string
	// Description is "<package> <version>".
	Description string
	// URL is the package's page on the registry.
	URL string
}

// RecordDeployment creates a deployment of the repository and marks it
// successful, so the Deployments sidebar lists the environment with the
// description and a link to URL. The call needs the deployments: write
// permission. It never waits for a build: the package is already published.
//
// Parameters: repo is the GitHub repository; d is what was published.
// Errors: the token may not create deployments (403), the tag may not exist on
// GitHub (422), or GitHub did not answer.
func (c *Client) RecordDeployment(repo Repository, d Deployment) error {
	var created struct {
		ID int64 `json:"id"`
	}
	// The release is already published, so no status check may hold it back, and
	// the deployment is not merged into a branch.
	body, _ := json.Marshal(map[string]any{
		"ref": d.Ref, "environment": d.Environment, "description": d.Description,
		"auto_merge": false, "required_contexts": []string{}, "production_environment": true,
	})
	if _, err := c.do(http.MethodPost, c.API+"/repos/"+repo.String()+"/deployments", "application/json", body, &created); err != nil {
		return err
	}
	status, _ := json.Marshal(map[string]any{
		"state": "success", "environment_url": d.URL, "description": d.Description,
	})
	_, err := c.do(http.MethodPost, fmt.Sprintf("%s/repos/%s/deployments/%d/statuses", c.API, repo.String(), created.ID), "application/json", status, nil)
	return err
}
