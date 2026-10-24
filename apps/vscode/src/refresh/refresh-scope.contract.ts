/**
 * How much of the view to read again: `affected` when only files changed
 * (which projects the changes touch), `full` when the projects themselves may
 * have changed (`smith.toml` edited, branch switched).
 */
export type RefreshScope = 'affected' | 'full'
