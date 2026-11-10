# doctor: written exception

- Path: apps/smith/doctor
- Rule waived: 12 production files is a review trigger (34 files).
- Constraint: the remaining files are the finding mappers and policies that one report run folds together (`run_doctor_use_case.go`). The checks that read the workspace for one concern each moved to `doctorchecks` and `lockfilechecks`. Dependency is one way: `doctor` imports `doctorchecks` and `lockfilechecks`, never the reverse. The rest still share the `Finding` and `Report` types of this slice.
- Owner: the smith maintainers.
- Temporary: yes. Move more mappers to `doctorchecks` as they stop using `Finding`.
