// Package pipelinetemplate is `smith pipeline`: ready-made pipeline files
// added to a workspace after its CI was generated (end-to-end tests, the
// dist/drop artifact, a GitHub Pages or Azure Functions deployment). They
// start the way every generated job does, and are written once: from then on
// they are the user's, and smith upgrade never rewrites them.
package pipelinetemplate
