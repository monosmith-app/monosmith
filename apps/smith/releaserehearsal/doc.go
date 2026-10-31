// Package releaserehearsal checks, before the first real release, that every
// registry a release publishes to is reachable, which credential it expects,
// that the configured credential is present and accepted, and that the CI
// identity may publish. It sends read-only requests and publishes nothing.
package releaserehearsal
