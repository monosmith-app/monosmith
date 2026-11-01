package releasing

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// distFile matches a distribution's file name on a simple index page and
// captures its version: name-1.2.3.tar.gz, name-1.2.3-py3-none-any.whl.
var distFile = regexp.MustCompile(`>[A-Za-z0-9_.]+?-(\d+(?:\.\d+)*)(?:\.tar\.gz|\.zip|-[^<>]*\.whl)<`)

// Published reads the simple index the project is uploaded to
// (TWINE_REPOSITORY_URL, else PyPI) and returns the newest version among its
// files. A project the index does not have is not published yet.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	project := params.Project.Name
	if len(params.Project.Exports) > 0 {
		project = params.Project.Exports[0]
	}
	index := targetOf(params.Project).Index
	result := protocol.PublishedResult{Registry: index}
	version, err := NewestOn(index, project)
	if err != nil {
		result.Unreachable = true
		result.Detail = protocol.LookupFailure(err)
		return result, nil
	}
	result.Version = version
	return result, nil
}

// NewestOn returns the newest release a simple index has of a distribution
// (pre-releases aside), "" when it has none.
func NewestOn(index, distribution string) (string, error) {
	page, err := httpGet(index + normalise(distribution) + "/")
	if err != nil {
		return "", err
	}
	newest := ""
	for _, m := range distFile.FindAllStringSubmatch(page, -1) {
		if newest == "" || newer(m[1], newest) {
			newest = m[1]
		}
	}
	return newest, nil
}

// newer compares dotted numeric versions.
func newer(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}
