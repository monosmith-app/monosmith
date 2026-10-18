package images

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// LegacyRegistryEnv is the environment variable that named the registry
// before [conventions.docker] registry; still read when the setting is unset.
const LegacyRegistryEnv = "CONTAINER_REGISTRY"

// Target is where release puts an app's image.
type Target struct {
	// Registries are where the image is pushed, in the order smith.toml
	// lists them; empty when none is configured.
	Registries []Registry
	// Image is the image's name under each registry.
	Image string
	// Push is false when the image is built but not pushed.
	Push bool
	// Legacy is true when the registry came from CONTAINER_REGISTRY.
	Legacy bool
	// DescriptionFrom is a workspace-relative file whose text becomes the
	// repository description on Docker Hub; empty for none.
	DescriptionFrom string
}

// Reference is the image's reference in a registry, without a tag
// (ghcr.io/acme/api).
func (t Target) Reference(r Registry) string { return r.URL + "/" + t.Image }

// LocalReference is the image's reference when it is only built, as image
// and verify do: the first registry's, so the name in the logs is the one
// release pushes, or the bare image name when there is no registry.
func (t Target) LocalReference() string {
	if len(t.Registries) > 0 {
		return t.Reference(t.Registries[0])
	}
	return t.Image
}

// Where names the registries, for messages (ghcr.io/acme, docker.io/acme).
func (t Target) Where() string {
	urls := make([]string, len(t.Registries))
	for i, r := range t.Registries {
		urls[i] = r.URL
	}
	return strings.Join(urls, ", ")
}

// TargetOf reads the project's docker settings ([conventions.docker]
// registry or registries and push, [projects.<name>.docker] image, push and
// their own registries). With no registry setting, CONTAINER_REGISTRY from
// env still names one. Push defaults to true when there is a registry. An
// error means the registry settings contradict each other.
func TargetOf(p protocol.Project, env func(string) string) (Target, error) {
	s := p.Settings["docker"]
	t := Target{Image: Name(p)}
	regs, legacy, err := RegistriesOf(s, env)
	if err != nil {
		return t, err
	}
	t.Registries, t.Legacy = regs, legacy
	if image, _ := s["image"].(string); image != "" {
		t.Image = image
	}
	t.Push = len(regs) > 0
	if push, ok := s["push"].(bool); ok {
		t.Push = push && len(regs) > 0
	}
	t.DescriptionFrom, _ = s["description_from"].(string)
	return t, nil
}

// Configured reports whether smith.toml names the registry (rather than
// CONTAINER_REGISTRY, or nothing).
func Configured(p protocol.Project) bool {
	s := p.Settings["docker"]
	registry, _ := s["registry"].(string)
	_, list := s["registries"]
	return registry != "" || list
}

// PushOff reports whether smith.toml turns the push off for the project.
func PushOff(p protocol.Project) bool {
	push, ok := p.Settings["docker"]["push"].(bool)
	return ok && !push
}
