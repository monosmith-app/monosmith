package pipeline

import (
	"regexp"
	"strings"
)

var (
	goDirective = regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)
	useBlock    = regexp.MustCompile(`(?ms)^use\s*\((.*?)\)`)
	useLine     = regexp.MustCompile(`(?m)^use\s+([^\s(]+)\s*$`)
)

// GoWork is what the plugin reads from go.work.
type GoWork struct {
	// GoVersion is major.minor, e.g. "1.24".
	GoVersion string
	// Use lists module directories, normalised to "dir/sub" (no "./").
	Use []string
}

// ParseGoWork reads the go directive and the use entries.
func ParseGoWork(src string) GoWork {
	var w GoWork
	if m := goDirective.FindStringSubmatch(src); m != nil {
		w.GoVersion = m[1]
	}
	add := func(entry string) {
		entry = strings.Trim(strings.TrimSpace(entry), `"`)
		if entry == "" || strings.HasPrefix(entry, "//") {
			return
		}
		w.Use = append(w.Use, strings.TrimPrefix(strings.ReplaceAll(entry, `\`, "/"), "./"))
	}
	for _, block := range useBlock.FindAllStringSubmatch(src, -1) {
		for _, line := range strings.Split(block[1], "\n") {
			line, _, _ = strings.Cut(line, "//")
			add(line)
		}
	}
	for _, m := range useLine.FindAllStringSubmatch(src, -1) {
		add(m[1])
	}
	return w
}
