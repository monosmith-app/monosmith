# convertplan: written exception

- Path: apps/smith/convertplan
- Rule waived: 12 production files is a review trigger (35 files).
- Constraint: the per-tool mappers (nx removal, rollup, esbuild, react, eslint, tag format) all append to the one `Plan` type and read the same Nx workspace model. A split would give each tool slice an import of the plan slice and the orchestrating `nx_removal_mapper.go` an import of each tool slice. That is acyclic but moves about 30 files for no new outcome.
- Owner: the smith maintainers.
- Temporary: no.
