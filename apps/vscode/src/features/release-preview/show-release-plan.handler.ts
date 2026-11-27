import type { CommandHost } from '../../shared/command-host/index.ts'
import { describeFailure } from '../../shared/smith-run/index.ts'
import { toTextDocument } from '../derive-files/index.ts'
import { say } from '../../shared/words/index.ts'

/**
 * Shows what a release would do as a page beside the code: `smith release --dry-run` (the
 * terminal's Dry run), or with `--notes` the release notes it would write. Nothing is published.
 * @param host - the folders and the Output channel of the extension
 * @param showPage - shows a page of text in the editor under a name
 * @param argument - what the editor passed to the command
 * @param notes - true for the release notes
 */
export async function showReleasePlan (host: CommandHost, showPage: (name: string, text: string) => Promise<void>, argument: unknown, notes: boolean): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  const args = ['release', '--dry-run', ...(notes ? ['--notes'] : [])]
  const result = await host.smithIn(state, true, 300_000)(args)
  const title = `$ smith ${args.join(' ')}`
  const text = result.code === 0 ? result.stdout : describeFailure(result)
  await showPage(notes ? 'release-notes.md' : 'release-dry-run.md', toTextDocument(say(notes ? 'releasepage.button.notes' : 'releasepage.button.preview'), `${title}\n\n${text}`))
}
