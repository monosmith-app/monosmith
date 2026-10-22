package components

import (
	"net/url"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PURL writes a package URL: pkg:<type>/<namespace>/<name>@<version>. The
// namespace and version may be empty; segments are percent-encoded.
func PURL(typ, namespace, name, version string) string {
	var b strings.Builder
	b.WriteString("pkg:" + typ + "/")
	if namespace != "" {
		b.WriteString(segment(namespace) + "/")
	}
	b.WriteString(segment(name))
	if version != "" {
		b.WriteString("@" + segment(version))
	}
	return b.String()
}

func segment(s string) string {
	e := url.PathEscape(s)
	e = strings.ReplaceAll(e, "@", "%40")
	return strings.ReplaceAll(e, "+", "%2B")
}

// New makes a component with its package URL.
func New(typ, namespace, name, version string) protocol.Component {
	full := name
	if namespace != "" {
		full = namespace + "/" + name
	}
	return protocol.Component{Name: full, Version: version, PURL: PURL(typ, namespace, name, version)}
}

// Dedupe drops repeated package URLs, keeping the first, and sorts by them.
func Dedupe(in []protocol.Component) []protocol.Component {
	seen := map[string]bool{}
	out := make([]protocol.Component, 0, len(in))
	for _, c := range in {
		if !seen[c.PURL] {
			seen[c.PURL] = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PURL < out[j].PURL })
	return out
}
