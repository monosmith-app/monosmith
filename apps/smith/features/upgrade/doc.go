// Package upgrade is `smith upgrade`: after installing a newer smith, bring the
// workspace up to it. smith.toml is migrated to the current schema, and the
// files smith owns (the pipelines) are regenerated, a stale one deleted; files
// smith wrote once (README, .gitignore) and the team's own are left alone.
package upgrade
