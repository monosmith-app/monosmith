package scaffold

import (
	"path"
	"strings"
)

// IsSecretName reports whether a file looks like it holds a secret: an
// environment file, a private key, a credentials file or a token store. smith
// never commits such a file without an explicit extra confirmation. The check
// reads the name only; it is a safety net, not a scanner.
//
// Parameters: rel is a slash-separated path relative to the folder.
// Returns: true when the base name matches a known secret pattern.
func IsSecretName(rel string) bool {
	name := strings.ToLower(path.Base(rel))
	switch {
	case name == ".env" || strings.HasPrefix(name, ".env."):
		for _, safe := range []string{".example", ".sample", ".template", ".dist"} {
			if strings.HasSuffix(name, safe) {
				return false
			}
		}
		return true
	case name == ".npmrc", name == ".pypirc", name == ".netrc", name == ".htpasswd":
		return true
	}
	for _, prefix := range []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa", "credentials", "secrets", "secret.", "service-account"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	for _, suffix := range []string{".pem", ".key", ".p12", ".pfx", ".keystore", ".jks"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
