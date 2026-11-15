package cipipeline

import "strings"

// GitHubCacheHygienePath is the workflow that keeps the Actions cache small.
const GitHubCacheHygienePath = ".github/workflows/cache-hygiene.yml"

// RenderGitHubCacheHygiene renders the workflow that deletes the caches no
// build will read again, so the 10 GB GitHub keeps for a repository holds the
// ones that matter: when a pull request closes, every cache of its merge ref;
// after the ci workflow on a branch, every smith cache of that branch but the
// newest of each name (a new lockfile saves a new npm cache and leaves the old
// one behind). Only keys that start with smith- are pruned on a branch; other
// actions manage their own.
//
// Parameters: m is the pipeline model (unused today: the workflow is the same
// for every workspace).
// Returns the workflow's YAML.
func RenderGitHubCacheHygiene(Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	b.WriteString(`name: cache hygiene

on:
  pull_request:
    types: [closed]
  workflow_run:
    workflows: [ci]
    types: [completed]

permissions:
  actions: write

jobs:
  prune:
    runs-on: ubuntu-latest
    env:
      GH_TOKEN: ${{ github.token }}
      GH_REPO: ${{ github.repository }}
    steps:
      - name: Delete the caches of the closed pull request
        if: github.event_name == 'pull_request'
        run: |
          gh cache list --ref "refs/pull/${{ github.event.pull_request.number }}/merge" --limit 100 --json id --jq '.[].id' |
            xargs -r -n1 gh cache delete
      - name: Keep only the newest smith cache of each name on the branch
        if: github.event_name == 'workflow_run'
        run: |
          gh cache list --ref "refs/heads/${{ github.event.workflow_run.head_branch }}" --limit 100 \
            --sort last_accessed_at --order desc --json id,key \
            --jq '[.[] | select(.key | startswith("smith-"))] | group_by(.key | sub("-[0-9a-f]{20,}$"; "")) | map(.[1:]) | add // [] | .[].id' |
            xargs -r -n1 gh cache delete
`)
	b.WriteString(slot("      ", "after-prune"))
	return b.String()
}
