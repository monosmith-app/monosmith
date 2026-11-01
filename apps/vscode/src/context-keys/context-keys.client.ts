import * as vscode from 'vscode'

/**
 * The context keys the extension sets, for `when` clauses of menus,
 * keybindings and welcome views: `monosmith.running` (a run is in progress),
 * `monosmith.state` (what the view shows: ready, starting, missing, untrusted,
 * failed, adoptable, empty, unrecognised or nofolder), `monosmith.repoKind` (what the adoptable folder is: nx, turborepo or repository), `monosmith.selectionKind` (the kind of row selected: task,
 * project, group or folder), `monosmith.hasLastRun` (a run was started
 * before, so it can be run again), `monosmith.affectedOnly` and
 * `monosmith.filtered` (the view is narrowed to the affected projects, or by a
 * text), `monosmith.fixable` (the doctor proposes fixes) and
 * `monosmith.derivedPaths` (the absolute paths of the derived files that need
 * action, which a menu tests with `resourcePath in monosmith.derivedPaths`).
 */
export type ContextKey = 'monosmith.running' | 'monosmith.state' | 'monosmith.repoKind' | 'monosmith.selectionKind' | 'monosmith.hasLastRun' | 'monosmith.affectedOnly' | 'monosmith.filtered' | 'monosmith.fixable' | 'monosmith.derivedPaths'

/** Sets a context key. */
export async function setContextKey (key: ContextKey, value: boolean | string | string[]): Promise<void> {
  await vscode.commands.executeCommand('setContext', key, value)
}
