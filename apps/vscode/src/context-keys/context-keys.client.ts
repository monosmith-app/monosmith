import * as vscode from 'vscode'

/**
 * The context keys the extension sets, for `when` clauses of menus,
 * keybindings and welcome views: `monosmith.running` (a run is in progress),
 * `monosmith.state` (what the view shows: ready, starting, missing, untrusted,
 * failed or empty), `monosmith.selectionKind` (the kind of row selected: task,
 * project, group or folder) and `monosmith.hasLastRun` (a run was started
 * before, so it can be run again).
 */
export type ContextKey = 'monosmith.running' | 'monosmith.state' | 'monosmith.selectionKind' | 'monosmith.hasLastRun'

/** Sets a context key. */
export async function setContextKey (key: ContextKey, value: boolean | string): Promise<void> {
  await vscode.commands.executeCommand('setContext', key, value)
}
