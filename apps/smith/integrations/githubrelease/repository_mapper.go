package githubrelease

import (
	"fmt"
	"strings"
)

// Repository is a GitHub owner/name pair.
type Repository struct {
	Owner string
	Name  string
}

// String renders owner/name.
func (r Repository) String() string { return r.Owner + "/" + r.Name }

// ParseRepository reads owner/name from a GitHub remote URL (https, ssh or
// scp-like) or from an "owner/name" string such as GITHUB_REPOSITORY.
func ParseRepository(remote string) (Repository, error) {
	s := strings.TrimSpace(remote)
	s = strings.TrimSuffix(strings.TrimSuffix(s, "/"), ".git")
	for _, prefix := range []string{"https://", "http://", "ssh://", "git://"} {
		if strings.HasPrefix(s, prefix) {
			s = strings.TrimPrefix(s, prefix)
			if at := strings.Index(s, "@"); at >= 0 && at < strings.Index(s+"/", "/") {
				s = s[at+1:]
			}
			_, s, _ = strings.Cut(s, "/")
			break
		}
	}
	if strings.HasPrefix(s, "git@") {
		_, s, _ = strings.Cut(s, ":")
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Repository{}, fmt.Errorf("cannot read a GitHub owner/repository from %q", remote)
	}
	return Repository{Owner: parts[0], Name: parts[1]}, nil
}
