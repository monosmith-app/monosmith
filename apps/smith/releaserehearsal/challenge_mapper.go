package releaserehearsal

import "strings"

// Challenge is what a feed's `WWW-Authenticate` header asks for: the scheme
// (Basic, Bearer, NuGet, Negotiate) and, for Basic, the realm.
type Challenge struct {
	Scheme string
	Realm  string
}

// ParseChallenge reads the first challenge of a `WWW-Authenticate` header
// value. An empty header gives the zero Challenge.
func ParseChallenge(header string) Challenge {
	header = strings.TrimSpace(header)
	if header == "" {
		return Challenge{}
	}
	scheme, rest, _ := strings.Cut(header, " ")
	c := Challenge{Scheme: scheme}
	if i := strings.Index(strings.ToLower(rest), "realm="); i >= 0 {
		realm := rest[i+len("realm="):]
		realm, _, _ = strings.Cut(realm, ",")
		c.Realm = strings.Trim(strings.TrimSpace(realm), `"`)
	}
	return c
}

// Expects says in words what credential the challenge asks for.
func (c Challenge) Expects() string {
	switch strings.ToLower(c.Scheme) {
	case "":
		return "no credential to read"
	case "basic":
		return "Basic (a user name and a token as the password)"
	case "bearer":
		return "Bearer (a token)"
	}
	return c.Scheme
}
