package convertplan

import (
	"regexp"
	"strings"
)

var (
	azurePoolName = regexp.MustCompile(`(?m)^pool:[ \t]*\r?\n(?:[ \t]+[^\r\n]*\r?\n)*?[ \t]+name:[ \t]*['"]?([^'"\s#]+)`)
	azureGroup    = regexp.MustCompile(`(?m)^[ \t]*-[ \t]*group:[ \t]*['"]?([^'"#\r\n]*?)['"]?[ \t]*(?:#[^\r\n]*)?\r?$`)
)

// AzurePipelineSettings reads what an existing Azure Pipelines definition
// runs on: the agent pool it names (a Microsoft-hosted image is smith's default
// and is not recorded) and the variable groups it links, which smith's
// generated definition would otherwise lose, with the secrets the groups hold.
func AzurePipelineSettings(text string) (pool string, groups []string) {
	if m := azurePoolName.FindStringSubmatch(text); m != nil {
		pool = m[1]
	}
	seen := map[string]bool{}
	for _, m := range azureGroup.FindAllStringSubmatch(text, -1) {
		name := strings.TrimSpace(m[1])
		if name != "" && !seen[name] {
			seen[name] = true
			groups = append(groups, name)
		}
	}
	return pool, groups
}
