package cipipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// SecretCredentials turns the names of [ci] secrets into credentials every
// step that runs smith receives, setup included: a build command or a private
// install may read them. None is required, so a local run does not need them.
func SecretCredentials(names []string) []protocol.Credential {
	out := make([]protocol.Credential, 0, len(names))
	for _, n := range names {
		out = append(out, protocol.Credential{Env: n, Description: "repository secret named in [ci] secrets", Setup: true, Tasks: true})
	}
	return out
}
