# cipipeline: written exception

- Path: apps/smith/cipipeline
- Rule waived: 12 production files is a review trigger (52 files).
- Constraint: the provider mappers (GitHub, GitLab, Azure, Bitbucket, CircleCI) all render the same `Model`, and the GitHub-named helpers sit inside shared files (`toolchain_setup_mapper.go`, `package_stores_mapper.go`, `release_assets_job_mapper.go`, `setup_steps_mapper.go`). A `githubworkflow` slice needs `Model` extracted into its own slice and those shared files split by provider first, or it would import `cipipeline` while `cipipeline` calls it.
- Owner: the smith maintainers.
- Temporary: yes. First step: move `pipeline_model.go` and `manual_release_model.go` to a model slice.
