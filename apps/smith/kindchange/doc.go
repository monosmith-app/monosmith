// Package kindchange is `smith kind`: turning a project into an app, an
// internal lib or a published package after it was created. The project's
// language plugin says which files change (its manifest, the references other
// projects and the workspace make to its folder); this package refuses what
// would break, moves the folder with git, writes the edits, and regenerates the
// pipelines.
package kindchange
