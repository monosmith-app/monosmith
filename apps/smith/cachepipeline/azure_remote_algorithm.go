package cachepipeline

import (
	"net/url"
	"strings"
)

// AzureRepo is where a git remote points inside Azure DevOps.
type AzureRepo struct {
	// Organization is the organisation's name, such as "acme".
	Organization string
	// Project is the project's name.
	Project string
	// Repository is the repository's name.
	Repository string
}

// OrganizationURL is the address az takes for --organization.
func (r AzureRepo) OrganizationURL() string { return "https://dev.azure.com/" + r.Organization }

// ParseAzureRemote reads the address of a git remote. It knows
// https://dev.azure.com/<org>/<project>/_git/<repo>, the ssh form
// git@ssh.dev.azure.com:v3/<org>/<project>/<repo>, and the older
// <org>.visualstudio.com forms. Any user name or password in the address is
// dropped and never returned.
//
// Parameters: raw is the output of `git remote get-url origin`.
// Returns: the parts, and false when the address is not an Azure DevOps one.
func ParseAzureRemote(raw string) (AzureRepo, bool) {
	raw = strings.TrimSpace(raw)
	var host, path string
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return AzureRepo{}, false
		}
		host, path = strings.ToLower(u.Hostname()), u.EscapedPath()
	} else if at := strings.Index(raw, "@"); at >= 0 && strings.Contains(raw[at:], ":") {
		rest := raw[at+1:]
		i := strings.Index(rest, ":")
		host, path = strings.ToLower(rest[:i]), "/"+rest[i+1:]
	} else {
		return AzureRepo{}, false
	}
	var parts []string
	for _, p := range strings.Split(strings.Trim(path, "/"), "/") {
		d, err := url.PathUnescape(p)
		if err != nil || d == "" {
			return AzureRepo{}, false
		}
		parts = append(parts, d)
	}
	switch {
	case host == "dev.azure.com":
		return fromGitPath(parts[0], parts[1:])
	case host == "ssh.dev.azure.com" || host == "vs-ssh.visualstudio.com":
		if len(parts) == 4 && parts[0] == "v3" {
			return AzureRepo{Organization: parts[1], Project: parts[2], Repository: parts[3]}, true
		}
	case strings.HasSuffix(host, ".visualstudio.com"):
		rest := parts
		if len(rest) > 0 && strings.EqualFold(rest[0], "DefaultCollection") {
			rest = rest[1:]
		}
		return fromGitPath(strings.TrimSuffix(host, ".visualstudio.com"), rest)
	}
	return AzureRepo{}, false
}

// fromGitPath reads <project>/_git/<repo>, or _git/<repo> where the project
// has the repository's name.
func fromGitPath(org string, rest []string) (AzureRepo, bool) {
	switch {
	case len(rest) == 3 && rest[1] == "_git":
		return AzureRepo{Organization: org, Project: rest[0], Repository: rest[2]}, true
	case len(rest) == 2 && rest[0] == "_git":
		return AzureRepo{Organization: org, Project: rest[1], Repository: rest[1]}, true
	}
	return AzureRepo{}, false
}
