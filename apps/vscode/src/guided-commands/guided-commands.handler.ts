import type { WebAppEnvironments } from '../workspace-view/index.ts'
import type { FolderState } from '../folder-session/index.ts'
import { askCommitBox } from '../commit-box/index.ts'
import type { CommandHost } from '../command-host/index.ts'
import { toTextDocument } from '../derive-files/index.ts'
import { describeFailure } from '../smith-run/index.ts'
import { environmentAnswers } from './environment-answers.mapper.ts'
import type { GuidedCommand } from './guided-command.contract.ts'
import { askGuidedStep, confirmGuided, tellDivergencesFailure, tellGuidedOutcome } from './guided-prompts.client.ts'
import { parseWebApps } from './guided-sources.mapper.ts'
import { runGuidedCommand } from './run-guided-command.use-case.ts'

/**
 * Runs a command of smith that has a screen of its own (environments, cache, pipeline...): asks, shows the output or the plan, and changes nothing before a yes.
 * @param host - the folders and the Output channel of the extension
 * @param showPage - shows a page of text in the editor under a name
 * @param spec - the command
 * @param argument - what the editor passed to the command
 */
export async function guidedHere (host: CommandHost, showPage: (name: string, text: string) => Promise<void>, spec: GuidedCommand, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  if (spec.mode === 'change' && !host.mayWrite(spec.title.replace(/…$/, '').toLowerCase())) return
  const outcome = await runGuidedCommand(spec, {
    run:       host.smithIn(state, true, 300_000),
    ask:       askGuidedStep,
    show:      async (title, text) => { await showPage(`${spec.id}.md`, toTextDocument(title, text)) },
    confirm:   confirmGuided,
    askCommit: async () => await askCommitBox(spec.title),
  }, environmentAnswers(spec.id, argument))
  host.report(spec.title, state, outcome)
  tellGuidedOutcome(spec.title, outcome, host.note)
  if (outcome.kind === 'done') await host.refreshFolder(state.root, 'full', false)
}

/**
 * Lists where the workspace chose differently from smith's defaults (`smith divergences`), with the reason and the way back.
 * @param host - the folders and the Output channel of the extension
 * @param showPage - shows a page of text in the editor under a name
 * @param argument - what the editor passed to the command
 */
export async function showDivergences (host: CommandHost, showPage: (name: string, text: string) => Promise<void>, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  const result = await host.smithIn(state)(['divergences'])
  if (result.code !== 0) {
    tellDivergencesFailure(describeFailure(result))

    return
  }
  await showPage('divergences.md', toTextDocument('Divergences from smith\'s defaults', result.stdout))
}

/**
 * The web apps of a folder and their environments, for the Projects view.
 * @param host - the folders and the Output channel of the extension
 * @param state - the folder
 * @returns the web apps, or none when smith lists none
 */
export async function readWebApps (host: CommandHost, state: FolderState): Promise<WebAppEnvironments[]> {
  try {
    const listed = await host.smithIn(state)(['environments', 'list', '--json'])

    return listed.code === 0 ? parseWebApps(listed.stdout) : []
  } catch {
    return []
  }
}
