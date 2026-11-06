import * as vscode from 'vscode'
import type { CommandHost } from '../command-host/index.ts'
import { chooseForced } from './forced-release.client.ts'
import { chooseBump, chooseHold, confirmRelease, tellReleaseOutcome } from './release-prompts.client.ts'
import { releaseWorkspace } from './release-workspace.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Releases the workspace: the plan, the packages to hold back, the bumps to change, then the last question.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 * @param released - runs when the release ended, so the Release Preview reads smith again
 */
export async function releaseHere (host: CommandHost, argument: unknown, released: () => Promise<void>): Promise<void> {
  const state = await host.folderFor(argument)
  if (state === undefined) return
  if (state.location === undefined) {
    void vscode.window.showWarningMessage(say('vscode.release-flow.monosmith-smith-was-not-found-for', { name: state.name }))

    return
  }
  if (!host.mayWrite('releasing')) return
  host.showOutput()
  const outcome = await releaseWorkspace({
    run:   host.smithIn(state),
    apply: async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: say('vscode.release-flow.monosmith-releasing-smith-tags-and-publishes', { name: state.name }) },
      async () => await host.smithIn(state, true, 1_800_000)(args),
    ),
    chooseHold,
    chooseBump,
    chooseForced,
    confirm: confirmRelease,
  })
  host.report('release', state, outcome)
  tellReleaseOutcome(outcome, host.note)
  await released()
}
