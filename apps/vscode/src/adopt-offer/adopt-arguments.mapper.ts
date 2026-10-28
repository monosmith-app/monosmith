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
