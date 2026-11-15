package cachesetup

import (
	"net/url"
	"regexp"
	"strings"
)

var unsafeSegment = regexp.MustCompile(`[^a-z0-9._-]+`)

// WorkspaceSegment turns a workspace name into a folder name: lowercase
// letters, digits, dots, dashes and underscores.
//
// Parameters: name is the workspace name, or its folder's name.
// Returns: the segment, "workspace" when nothing is left.
func WorkspaceSegment(name string) string {
	seg := strings.Trim(unsafeSegment.ReplaceAllString(strings.ToLower(name), "-"), "-.")
	if seg == "" {
		return "workspace"
	}
	return seg
}

// JoinPrefix joins folder names with single slashes and no slash at the ends.
func JoinPrefix(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p = strings.Trim(p, "/"); p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, "/")
}

// AzureURL is the remote of a container and a prefix.
func AzureURL(account, container, prefix string) string {
	return "azblob://" + JoinPrefix(account, container, prefix)
}

// S3URL is the remote of a bucket and a prefix.
func S3URL(bucket, prefix string) string {
	return "s3://" + JoinPrefix(bucket, prefix)
}

// ValidateHTTPS checks a typed address.
//
// Returns: "" when it is usable, else what is wrong, in a sentence.
func ValidateHTTPS(address string) string {
	u, err := url.Parse(address)
	switch {
	case err != nil || u.Host == "":
		return "type a full address such as https://cache.acme.dev/smith"
	case u.Scheme != "https" && u.Scheme != "http":
		return "the address must start with https:// (or http:// for a server on a trusted network)"
	case u.User != nil:
		return "leave the user name and password out of the address: a token goes in SMITH_CACHE_TOKEN"
	}
	return ""
}
