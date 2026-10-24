package configedit

import "strings"

const defaultMarker = "Default: "

// Default is the value a setting takes when smith.toml does not set it, as
// the schema's description states it ("Default: false."); empty when the
// description does not say.
func (f Field) Default() string {
	i := strings.LastIndex(f.Description, defaultMarker)
	if i < 0 {
		return ""
	}
	return strings.TrimSuffix(strings.TrimSpace(f.Description[i+len(defaultMarker):]), ".")
}

// Explanation is what the setting does, without the closing "Default: ..."
// sentence that Default reports.
func (f Field) Explanation() string {
	i := strings.LastIndex(f.Description, defaultMarker)
	if i < 0 {
		return f.Description
	}
	return strings.TrimSpace(f.Description[:i])
}
