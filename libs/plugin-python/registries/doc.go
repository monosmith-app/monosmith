// Package registries points pip and twine at the private registries
// smith.toml names: the index pip installs from (the venv's site
// configuration, written by CI setup), Azure Pipelines' PipAuthenticate@1 and
// TwineAuthenticate@1 sign-ins, and the credentials twine uploads with.
package registries
