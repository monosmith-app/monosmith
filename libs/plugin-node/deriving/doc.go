// Package deriving holds the Node renderers of derived files (ADR 0046): the
// references of a tsconfig, from the internal dependencies of the project
// graph, and the display name of a Jest config, from the package name. A
// renderer answers with content; the host owns the writes and the record.
package deriving
