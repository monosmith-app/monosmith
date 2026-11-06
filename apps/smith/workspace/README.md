# workspace: written exception

- Path: apps/smith/workspace
- Rule waived: 12 production files is a review trigger (31 files).
- Constraint: every file is part of one outcome, reading and checking smith.toml. The validators all take the `Config` type this slice owns, so a split would make the new slices import `workspace` while `workspace` calls them (a cycle). The parts that did not need `Config` already moved out: `releasetable` (release table resolution), `releasetag` (tag format) and `dockerregistries` (registry and context checks).
- Owner: the smith maintainers.
- Temporary: no. Revisit if `Config` is split into a contract slice.
