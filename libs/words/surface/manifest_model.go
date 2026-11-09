package surface

import "strings"

// Manifest is what one surface claims: exact keys, and prefixes that claim every
// key that starts with one of them ("vscode." claims every vscode.* key).
type Manifest struct {
	// Keys are exact catalogue keys the surface uses.
	Keys []string `json:"keys"`
	// Prefixes are key prefixes, each claiming every key that starts with it.
	Prefixes []string `json:"prefixes"`
}

// Claims reports whether the manifest claims key, by exact match or by prefix.
func (m Manifest) Claims(key string) bool {
	for _, k := range m.Keys {
		if k == key {
			return true
		}
	}
	for _, p := range m.Prefixes {
		if p != "" && strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
