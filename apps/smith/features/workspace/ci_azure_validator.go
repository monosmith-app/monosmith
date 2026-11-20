package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

// buildNumberPlaceholder finds the {name} placeholders of a build number pattern.
var buildNumberPlaceholder = regexp.MustCompile(`\{[^}]*\}`)

// buildNumberForbidden are the characters Azure Pipelines refuses in a build number.
const buildNumberForbidden = `"/:<>\|?@*`

// validateCIAzure checks [ci.azure]: build_tags is true, false or a table of
// project names to tag names; build_number is true, false or a pattern that
// uses only {project} and {version} and no character Azure refuses.
func validateCIAzure(a AzureSection) error {
	switch v := a.BuildTags.(type) {
	case nil, bool:
	case map[string]any:
		for project, name := range v {
			s, ok := name.(string)
			if !ok || strings.TrimSpace(s) == "" {
				return fmt.Errorf("%s: [ci.azure] build_tags.%s must be the tag's name, a non-empty string", FileName, project)
			}
			if strings.ContainsAny(s, ";\r\n") {
				return fmt.Errorf("%s: [ci.azure] build_tags.%s must not hold a semicolon or a line break, not %q", FileName, project, s)
			}
		}
	default:
		return fmt.Errorf("%s: [ci.azure] build_tags must be true, false or a table of tag names ({ api = \"api-release\" })", FileName)
	}
	switch v := a.BuildNumber.(type) {
	case nil, bool:
	case string:
		for _, ph := range buildNumberPlaceholder.FindAllString(v, -1) {
			if ph != "{project}" && ph != "{version}" {
				return fmt.Errorf("%s: [ci.azure] build_number has unknown placeholder %s (use {project} and {version})", FileName, ph)
			}
		}
		if strings.ContainsAny(v, buildNumberForbidden+"\r\n") {
			return fmt.Errorf("%s: [ci.azure] build_number must not hold any of %s, which Azure Pipelines refuses in a build number, not %q", FileName, buildNumberForbidden, v)
		}
	default:
		return fmt.Errorf("%s: [ci.azure] build_number must be a pattern such as \"{project}_{version}\", or false", FileName)
	}
	return nil
}
