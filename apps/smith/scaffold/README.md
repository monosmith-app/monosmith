# scaffold: written exception

- Path: apps/smith/scaffold
- Rule waived: 12 production files is a review trigger (41 files).
- Constraint: `smith new` and `smith add` share the question, plan and selection files (`plan_selection`, `plan_recipe`, `choose_language`, `step_times`, `record_environments`, `ValidateName`, `nodeWorkspaceDirectory`). Splitting the add outcome out needs those shared files in their own slice first, and about a dozen unexported symbols exported. Removal already moved to `removeproject`.
- Owner: the smith maintainers.
- Temporary: yes. Extract the shared plan and ask files, then split new and add.
