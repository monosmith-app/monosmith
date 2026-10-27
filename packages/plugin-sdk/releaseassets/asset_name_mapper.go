package releaseassets

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var placeholder = regexp.MustCompile(`\{(\w+)\}`)

// DefaultName is the file name pattern when the spec gives none.
const DefaultName = "{product}_{version}_{os}_{arch}.{ext}"

// AssetName fills an asset name pattern for one platform. goos and goarch are
// Go's names; the spec's aliases rename them in the file name only.
func AssetName(spec protocol.AssetSpec, project, version, goos, goarch, ext string) (string, error) {
	product := spec.Product
	if product == "" {
		product = project
	}
	if a, ok := spec.OSAlias[goos]; ok {
		goos = a
	}
	if a, ok := spec.ArchAlias[goarch]; ok {
		goarch = a
	}
	values := map[string]string{"product": product, "version": version, "os": goos, "arch": goarch, "ext": ext}
	pattern := spec.Name
	if pattern == "" {
		pattern = DefaultName
	}
	var unknown []string
	name := placeholder.ReplaceAllStringFunc(pattern, func(m string) string {
		v, ok := values[m[1:len(m)-1]]
		if !ok {
			unknown = append(unknown, m)
		}
		return v
	})
	if len(unknown) > 0 {
		return "", fmt.Errorf("asset name %q has unknown placeholder %s; use {product}, {version}, {os}, {arch} or {ext}", pattern, strings.Join(unknown, ", "))
	}
	return name, nil
}
