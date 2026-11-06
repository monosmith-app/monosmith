import * as vscode from 'vscode'
import { chooseFolder, type CommandHost } from '../command-host/index.ts'
import { pickNewQuestion } from '../new-workspace/index.ts'
import { smithForTerminal } from '../smith-terminal/index.ts'
import { askConvertCommit, confirmConvert, tellConvertOutcome, tellNoConvertible } from './convert-prompts.client.ts'
import { convertWorkspace } from './convert-workspace.use-case.ts'
import { say } from '../words/index.ts'

/** A folder that looks like a repository smith can bring in. */
export interface ConvertOffer {
  root: string
  name: string
  kind: string
}

/**
 * Converts an Nx or Turborepo workspace with the questions of `smith convert --questions --json`:
 * the level, and at the full level the four choices. Nothing is changed before the plan is shown
 * and the person has said yes; smith commits only when the box is ticked.
 * @param host - the folders and the Output channel of the extension
 * @param offers - the folders without a smith.toml that look like a repository smith recognises
 * @param converted - runs when the workspace is converted, so the view reads it
 */
export async function convertChoices (host: CommandHost, offers: readonly ConvertOffer[], converted: () => Promise<void>): Promise<void> {
  const chosen = await chooseFolder(offers.filter((offer) => offer.kind === 'nx' || offer.kind === 'turborepo'))
  if (chosen === undefined) {
    tellNoConvertible()

    return
  }
  const executable = await smithForTerminal(chosen.root)
  if (executable === undefined) return
  host.showOutput()
  const inFolder = host.smithAt(executable, chosen.root, chosen.name, 600_000)
  const outcome = await convertWorkspace({
    run:   inFolder,
    apply: async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: say('vscode.convert-workspace.monosmith-converting', { name: chosen.name }) },
      async () => await inFolder(args),
    ),
    pick:      pickNewQuestion,
    askCommit: askConvertCommit,
    confirm:   confirmConvert,
  })
  tellConvertOutcome(outcome, host.note)
  if (outcome.kind === 'converted') await converted()
}
