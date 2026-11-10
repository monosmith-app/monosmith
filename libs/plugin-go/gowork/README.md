# gowork

Reads and edits `go.work` and `go.mod`: parse the use entries and the go directive, find the module of a project, add or drop a use entry, and remove a module from `go.work` (`smith remove`). It imports no sibling slice. `pipeline` and `dependencies` depend on it.
