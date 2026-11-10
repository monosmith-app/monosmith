# ADR 0080: A manual release is a presented choice, and CI offers it only on request

- Status: accepted
- Date: 2026-11-02

## Context

A user (MVD) had a package whose last release had no fix in it: the registry package was fine, but
the person wanted the same code under a new version (a re-publish after a registry problem, a
dependency range that must move). No commit in the package released it, so `smith release --bump
mvd=patch` refused with "mvd is not in this plan, so there is nothing to change", and the CI had no
way to start such a release. The only way out was an empty `fix:` commit, which puts a lie in the
history.

Publishing and tagging cannot be undone, and in CI whoever can run a workflow can publish. A manual
release therefore must be a choice that is presented with its consequences, never a default.

## Decision

- `smith release --bump <project>=<level> --force` adds a package that no commit releases to the
  plan as a manual release. The plan says: "no commit since mvd@0.14.0 releases this: you are
  releasing the same code under a new version". The final question says what happens: a new tag at
  HEAD, a registry publish, a GitHub Release and the assets, and that none of it can be undone.
- Without a terminal the release needs `--yes` as well. `--force` without `--bump` is refused and
  prints the usage. For a package that already has a releasing commit, `--force` only applies the
  bump. A release channel forces the next number of the channel (ADR 0051).
- The report (text, markdown, JSON `manual`) and the tag message record that the release was manual.
- `[ci] manual_release = true` adds the input to the generated pipelines. It is off by default.
  GitHub gets a `workflow_dispatch` input `release_bump` (`project=patch|minor|major`, empty by
  default), passed to `smith ci release --bump ... --force` through an environment variable, never
  inside the script. The ci job, the native job and the assets jobs run for a manual release as for
  a push. The step summary says the release was manual, the value and who ran it.
- Azure gets a runtime parameter, GitLab a variable of the "Run pipeline" form, Bitbucket a custom
  pipeline with a variable. CircleCI has no form for a value, so `smith ci generate` says so in a
  note and the docs name the alternatives.
- The terminal UI and the VS Code release command offer the path as a presented choice for a package
  that is not in the plan, with the consequences, and run the same flags.

## Consequences

- The history stays true: no empty commit pretends to be a fix.
- A workspace that does not opt in has no new input, and the generated files do not change.
- A manual release does not run the pack step on Azure, GitLab and Bitbucket (it runs on GitHub).
- A package that depends on a force-released package is not released with it: the code did not
  change, so its published range stays correct.

## Where to check

`apps/smith/release/force_release_use_case.go` and its tests, `apps/smith/cipipeline/manual_release_*`
and the golden file `testdata/github_manual_release.golden.yml`.
