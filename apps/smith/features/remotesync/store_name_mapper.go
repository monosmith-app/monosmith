package remotesync

import (
	"net/url"
	"strings"
)

// StoreName is a remote cache URL as smith prints it: the scheme, the host and
// the prefix, never credentials in the address or a query (which can hold a
// token).
//
// Parameters: remote is the configured URL.
// Returns: "s3://bucket/prefix"; the input cut at "?" when it does not parse.
func StoreName(remote string) string {
	u, err := url.Parse(remote)
	if err != nil {
		before, _, _ := strings.Cut(remote, "?")
		return before
	}
	name := u.Scheme + "://" + u.Host
	if prefix := strings.Trim(u.Path, "/"); prefix != "" {
		name += "/" + prefix
	}
	return name
}
