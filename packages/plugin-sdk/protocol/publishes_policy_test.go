package protocol

import "testing"

func TestAPackageThatPublishesToNoRegistryDoesNotPublish(t *testing.T) {
	for name, tc := range map[string]struct {
		p    Project
		want bool
	}{
		"a package":                {Project{Kind: ProjectPackage}, true},
		"an app":                   {Project{Kind: ProjectApp}, false},
		"a package with registry":  {Project{Kind: ProjectPackage, Registries: &Registries{Publish: &Registry{Name: "internal"}}}, true},
		"a package, registry none": {Project{Kind: ProjectPackage, Registries: &Registries{NoPublish: true}}, false},
	} {
		if got := tc.p.Publishes(); got != tc.want {
			t.Errorf("%s: Publishes() = %v, want %v", name, got, tc.want)
		}
	}
}
