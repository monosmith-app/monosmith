# pipeline

What the Go plugin tells CI, `smith doctor` and `smith remove` about a Go workspace.

## Written exception

| Path | Rule waived | Constraint | Owner | Temporary |
|---|---|---|---|---|
| `libs/plugin-go/pipeline` (14 production files) | the 12-file review trigger | No clean seam splits it yet. The CI contribution (`ci_contribution_use_case.go`) and the doctor checks share `go_work_mapper.go`, `module_file_mapper.go` and `toolchainVersion`, and `remove_module_use_case.go` and the `go.work` editors use the same parser. A split needs a new slice for `go.work` (parse, add, drop, remove) that both depend on. The Node plugin split its `ci` slice off on the same pattern. | the smith maintainers | yes: split into `ci`, `gowork` and the checks when the next change touches two of these groups |
