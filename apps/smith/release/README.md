# release: written exception

- Path: apps/smith/release
- Rule waived: 12 production files is a review trigger (44 files).
- Constraint: planning, tagging, publishing and the journal all use the `Item`, `Plan` and `Report` types. `run_release_use_case.go` was cut from 770 to under 470 lines by moving its pure and single-step parts beside it. The registry and lineage checks already live in `releaselineage`, `releasepack` and `releasepreview`.
- Owner: the smith maintainers.
- Temporary: yes. Split planning from publishing once the shared types have their own slice.
