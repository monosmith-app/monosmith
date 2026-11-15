package updatenotice

import (
	"context"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/smithreleases"
)

// ReleasesURL is where the release notes are.
const ReleasesURL = "https://github.com/monosmith-app/monosmith/releases"

// FetchLatest asks smith's releases for the newest version on a channel,
// giving up after timeout. It is the one request smith makes for the update
// notice: GitHub's list of smith releases, or the mirror's latest.txt and
// next.txt (SMITH_DOWNLOAD_URL). Nothing about the person or the workspace is
// sent. The npm package is a launcher with its own version line now (ADR
// 0101), so its dist-tags no longer name smith's versions.
//
// Parameters: channel is "latest" or "next".
// Returns: the newest version; an error when no release answers in time.
func FetchLatest(ctx context.Context, timeout time.Duration, channel string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return smithreleases.Newest(ctx, channel)
}
