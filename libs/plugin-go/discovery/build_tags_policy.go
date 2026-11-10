package discovery

import (
	"fmt"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var buildTagPattern = regexp.MustCompile(`^[A-Za-z0-9_.]+$`)

// BuildTags returns the Go build tags a project's settings ask for: "tags",
// then "tags_<os>" for the OS smith runs on (linux, windows or darwin),
// without duplicates. The settings are the project's
// [projects.<name>.settings.go] table. A value that is not a list of tag names
// (letters, digits, "_" and ".") is an error that names the key.
func BuildTags(settings map[string]any) ([]string, error) {
	return buildTagsFor(settings, runtime.GOOS)
}

func buildTagsFor(settings map[string]any, goos string) ([]string, error) {
	var tags []string
	for _, key := range []string{"tags", "tags_" + goos} {
		raw, ok := settings[key]
		if !ok {
			continue
		}
		list, err := tagList(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		for _, t := range list {
			if !slices.Contains(tags, t) {
				tags = append(tags, t)
			}
		}
	}
	return tags, nil
}

func tagList(raw any) ([]string, error) {
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("must be a list of build tags, such as [\"production\", \"gtk3\"]")
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		s, ok := it.(string)
		if !ok || !buildTagPattern.MatchString(s) {
			return nil, fmt.Errorf("%v is not a build tag: use letters, digits, \"_\" and \".\" in each tag", it)
		}
		out = append(out, s)
	}
	return out, nil
}

// withBuildTags passes tags to the go commands of tasks as -tags=a,b, and to
// golangci-lint as --build-tags a,b. The command is part of the task's cache
// key, so changing the tags misses the cache. deps-check does not compile and
// is left alone.
func withBuildTags(tasks map[string]protocol.Task, tags []string) {
	if len(tags) == 0 {
		return
	}
	csv := strings.Join(tags, ",")
	for name, t := range tasks {
		switch name {
		case "build", "test", "package", "start":
			t.Command = slices.Insert(slices.Clone(t.Command), 2, "-tags="+csv)
		case "lint":
			t.Command = append(slices.Clone(t.Command), "--build-tags", csv)
		default:
			continue
		}
		tasks[name] = t
	}
}

// projectBuildTags reads the tags of the project called name from the
// settings the host sent, by project name.
func projectBuildTags(settings map[string]map[string]any, name string) ([]string, error) {
	tags, err := BuildTags(settings[name])
	if err != nil {
		return nil, fmt.Errorf("smith.toml [projects.%s.settings.go] %w", name, err)
	}
	return tags, nil
}
