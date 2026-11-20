// Package projectimport is `smith import <git-url|folder>`: it brings an
// existing project into the smith workspace that already exists (ADR 0106).
// The steps are the ones adopt takes for the project at a repository's root:
// fetch (a folder as it is, or a clone of a git repository), identify (the
// language plugins' manifest claims, through adoptmove.Detect), place (git
// history is merged under the folder, or squashed with git subtree add; a
// plain folder is copied),
// split (adoptsplit: the tools go to the workspace, the nested lockfile is
// dropped), join (the workspace/join protocol method), then the pipelines,
// smith install and smith doctor. smith never picks the folder or the
// history for the person: a terminal asks, a flag says it, or smith refuses.
package projectimport
