package plugininstall

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var repoPart = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// ParseSource reads an install argument: github:owner/repo@tag[#asset], an
// https URL, or a path. A tag is required: installing "whatever is latest"
// would make the digest the only thing that says what was installed.
func ParseSource(arg string) (Source, error) {
	switch {
	case strings.HasPrefix(arg, "github:"):
		return parseGitHub(strings.TrimPrefix(arg, "github:"))
	case strings.HasPrefix(arg, "https://"), strings.HasPrefix(arg, "http://"):
		if err := CheckURL(arg); err != nil {
			return Source{}, err
		}
		return Source{Kind: KindURL, Location: arg}, nil
	case strings.Contains(arg, "://"):
		return Source{}, fmt.Errorf("%s: only https URLs, github:owner/repo@tag and local paths are accepted", arg)
	case strings.TrimSpace(arg) == "":
		return Source{}, fmt.Errorf("say what to install: a path, an https URL or github:owner/repo@tag")
	}
	return Source{Kind: KindLocal, Location: arg}, nil
}

func parseGitHub(rest string) (Source, error) {
	asset := ""
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		rest, asset = rest[:i], rest[i+1:]
	}
	repo, tag, hasTag := strings.Cut(rest, "@")
	owner, name, ok := strings.Cut(repo, "/")
	if !ok || !repoPart.MatchString(owner) || !repoPart.MatchString(name) {
		return Source{}, fmt.Errorf("%q is not github:owner/repo@tag", "github:"+rest)
	}
	if !hasTag || strings.TrimSpace(tag) == "" {
		return Source{}, fmt.Errorf("name the release to install: github:%s/%s@<tag>", owner, name)
	}
	if strings.ContainsAny(asset, "/\\") {
		return Source{}, fmt.Errorf("asset %q must be a file name", asset)
	}
	return Source{Kind: KindGitHub, Owner: owner, Repo: name, Tag: tag, Asset: asset}, nil
}

// CheckURL accepts an https URL, and an http one only for this machine
// (localhost, a loopback address), so nothing travels in the clear that could
// be changed on the way.
func CheckURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fmt.Errorf("%q is not a URL", raw)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if host == "localhost" {
			return nil
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return nil
		}
	}
	return fmt.Errorf("%s: only https URLs are accepted (http only for localhost)", raw)
}
