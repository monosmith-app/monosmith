// Package releaseassets is `smith release assets` and the `assets` phase of
// smith ci: on one runner of a platform it runs the build command of every
// project released at HEAD whose [release] has one, and uploads the files the
// build leaves to the project's GitHub Release with gh release upload.
package releaseassets
