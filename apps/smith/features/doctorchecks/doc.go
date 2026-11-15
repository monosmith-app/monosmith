// Package doctorchecks holds the doctor checks that read the workspace for one
// concern each (build outputs, task folding, Tauri sidecars, web environments,
// recipes), and the vocabulary every finding shares: its areas, how a problem
// is worded and how a long list is summed up. The doctor slice runs them and
// depends on this one, never the other way round.
package doctorchecks
