// Package pluginbundle is the bundle plugin for smith: a directory with an
// app.json and a flow.json and no other manifest is a project that is not
// built from a language's sources (an app assembled from other projects'
// builds). Its version is app.json's, its release is a git tag. Its code lives
// in the slice packages below; plugin is the entry point.
package pluginbundle
