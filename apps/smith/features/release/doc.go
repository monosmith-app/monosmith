// Package release versions, publishes and tags every publishable package from
// its conventional commits. The model is tag-only: manifests are stamped for
// the build and publish, then restored, and the tag is the record of the
// release. Packages release in dependency order, and a re-run after a failure
// resumes: what is tagged is done, and a version the registry already has is
// skipped.
package release
