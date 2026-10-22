package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's gradle.lockfile, at the project or any directory
// above it up to the workspace, and falls back to its pom.xml, then its
// Gradle build script.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "gradle.lockfile")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		return protocol.ComponentsResult{Components: ComponentsFromGradleLock(string(lock.Body)), Source: lock.Rel}, nil
	}
	if pom, ok, err := components.ReadIn(root, params.Project.Root, "pom.xml"); err != nil {
		return protocol.ComponentsResult{}, err
	} else if ok {
		list, err := ComponentsFromPOM(pom.Body)
		return protocol.ComponentsResult{Components: list, Source: pom.Rel}, err
	}
	for _, name := range []string{"build.gradle.kts", "build.gradle"} {
		script, ok, err := components.ReadIn(root, params.Project.Root, name)
		if err != nil {
			return protocol.ComponentsResult{}, err
		}
		if ok {
			return protocol.ComponentsResult{Components: ComponentsFromGradleScript(string(script.Body)), Source: script.Rel}, nil
		}
	}
	return protocol.ComponentsResult{}, nil
}
