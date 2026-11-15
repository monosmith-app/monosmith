package packagemanifests

// Release is what the manifests describe.
type Release struct {
	// Product names the package; Binary the executable in each archive
	// (without ".exe").
	Product, Binary string
	Version         string
	Description     string
	Homepage        string
	License         string
	// Owner and Repo are the GitHub repository the assets download from.
	Owner, Repo string
	Tag         string
	Archives    []Archive
}

// Archive is one platform's zip.
type Archive struct {
	Name   string
	OS     string
	Arch   string
	SHA256 string
}

// URL is where an archive downloads from.
func (r Release) URL(a Archive) string {
	return "https://github.com/" + r.Owner + "/" + r.Repo + "/releases/download/" + escapeTag(r.Tag) + "/" + a.Name
}

func escapeTag(tag string) string {
	out := ""
	for _, c := range tag {
		switch c {
		case '@':
			out += "%40"
		case '/':
			out += "%2F"
		default:
			out += string(c)
		}
	}
	return out
}

func (r Release) find(os, arch string) (Archive, bool) {
	for _, a := range r.Archives {
		if a.OS == os && a.Arch == arch {
			return a, true
		}
	}
	return Archive{}, false
}
