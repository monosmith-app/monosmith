package registries

import (
	"net/url"
	"strings"
)

// AzureFeed is how Azure Pipelines' authenticate tasks name an Azure
// Artifacts feed: "project/feed" for a project-scoped feed, "feed" for an
// organisation's, from a URL such as
// https://pkgs.dev.azure.com/acme/web/_packaging/libs/pypi/simple/.
// "" when the URL is not an Azure Artifacts feed.
func AzureFeed(feedURL string) string {
	u, err := url.Parse(feedURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, p := range parts {
		if p != "_packaging" || i+1 >= len(parts) {
			continue
		}
		feed := parts[i+1]
		// pkgs.dev.azure.com/<org>/<project>/_packaging/...: the project is
		// the segment before _packaging, after the organisation.
		if strings.EqualFold(u.Host, "pkgs.dev.azure.com") && i == 2 {
			return parts[1] + "/" + feed
		}
		// <org>.pkgs.visualstudio.com/<project>/_packaging/...
		if strings.HasSuffix(strings.ToLower(u.Host), ".pkgs.visualstudio.com") && i == 1 {
			return parts[0] + "/" + feed
		}
		return feed
	}
	return ""
}
