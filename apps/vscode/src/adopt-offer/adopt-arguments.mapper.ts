/** What a click on the welcome view asks for. */
export type AdoptAction = 'start' | 'preview'

/**
 * The arguments to run smith with in a terminal. `start` is plain `smith`,
 * which opens its screen on a folder without a smith.toml: it says what it
 * found, asks how far to go and lists every file before it changes any.
 * `preview` is `smith adopt --dry-run`, which only reports (an Nx or
 * Turborepo workspace is handed to `smith convert --dry-run`).
 */
export function toAdoptArguments (action: AdoptAction): string[] {
  return action === 'preview' ? ['adopt', '--dry-run'] : []
}

/**
 * The arguments for a project at the repository root once the person said
 * where it goes: `smith adopt --init --move <folder>`, which never asks. `none`
 * keeps the project where it is. With `commit`, `--commit` makes the move a
 * commit of its own and the adoption one more that `smith undo` reverts;
 * without it smith commits nothing.
 * @param dir - the folder, or `none`
 * @param commit - whether the person ticked "Commit this change so it can be undone"
 */
export function toMoveArguments (dir: string, commit: boolean): string[] {
  return ['adopt', '--init', '--move', dir, ...(commit ? ['--commit'] : [])]
}
