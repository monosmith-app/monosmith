import * as vscode from 'vscode'
import { chooseFolder, type CommandHost } from '../command-host/index.ts'
import type { BlankFolder } from '../empty-folder/index.ts'
import { openSmithTerminal, smithForTerminal } from '../smith-terminal/index.ts'
import { toAdoptArguments, toMoveArguments, type AdoptAction } from './adopt-arguments.mapper.ts'
import type { RepositoryOffer } from './repository-offer.contract.ts'
import { LEAVE } from './move-choices.policy.ts'
import { askWhatToSplit } from './where-to-split.client.ts'
import { askWhereToMove } from './where-to-move.client.ts'
import { say } from '../words/index.ts'

/** The folders without a smith.toml that smith can bring in. */
export interface AdoptCandidates {
  /** Folders that look like a repository smith recognises. */
  offers: readonly RepositoryOffer[]
  /** Folders with nothing in them, and folders with files smith does not recognise. */
  blanks: readonly BlankFolder[]
}

/**
 * Offers Undo when the command in the terminal ends well. It needs the
 * shell integration of the terminal; without it the offer is not made, and
 * "Undo the last change" stays in the palette.
 */
function offerUndoWhenDone (terminal: vscode.Terminal, what: string, subscriptions: vscode.Disposable[]): void {
  const ended = vscode.window.onDidEndTerminalShellExecution((event) => {
    if (event.terminal !== terminal) return
    ended.dispose()
    if (event.exitCode !== 0) return
    void vscode.window.showInformationMessage(say('vscode.adopt-offer.monosmith-if-it-is-not-what', { what }), say('commands.describe.undo.title')).then(async (choice) => {
      if (choice !== undefined) await vscode.commands.executeCommand('monosmith.undo')
    })
  })
  subscriptions.push(ended)
}

/**
 * Opens smith in a terminal on the folder to bring in: the repositories it
 * recognises, else the folders with files. Nothing is written by this
 * click: smith asks first.
 * @param host - the folders and the Output channel of the extension
 * @param found - the folders without a smith.toml
 * @param subscriptions - the lifetime the Undo offer ends with
 * @param action - `start` brings the folder in, `preview` only shows what would happen
 */
export async function adopt (host: CommandHost, found: AdoptCandidates, subscriptions: vscode.Disposable[], action: AdoptAction): Promise<void> {
  const candidates = found.offers.length > 0
    ? found.offers.map((offer) => ({ root: offer.root, name: offer.name, why: `found ${offer.marker}`, project: offer.project }))
    : found.blanks.filter((folder) => folder.kind === 'unrecognised').map((folder) => ({ root: folder.root, name: folder.name, why: say('vscode.adopt-offer.no-tool-it-recognises'), project: undefined }))
  const chosen = await chooseFolder(candidates)
  if (chosen === undefined) return
  // One project at the root: ask where it goes (smith never picks), then run the step that does not ask again.
  const answer = action === 'start' && chosen.project !== undefined ? await askWhereToMove(chosen.project) : undefined
  if (answer === undefined && action === 'start' && chosen.project !== undefined) return
  const executable = await smithForTerminal(chosen.root)
  if (executable === undefined) return
  // A project that moves is split from the workspace: ask what smith cannot place (nothing is changed yet).
  const moving = answer !== undefined && answer.dir !== LEAVE
  const split = moving ? await askWhatToSplit(executable, chosen.root, answer.dir) : undefined
  if (moving && split === undefined) return
  host.note(say('vscode.adopt-offer.monosmith-into-monosmith', { part: say(action === 'start' ? 'vscode.adopt-offer.bringing' : 'vscode.adopt-offer.previewing-bringing'), name: chosen.name, why: chosen.why }))
  const terminal = openSmithTerminal(executable, answer === undefined ? toAdoptArguments(action) : toMoveArguments(answer.dir, answer.commit, split), chosen.root)
  if (action === 'start' && (answer === undefined || answer.commit)) offerUndoWhenDone(terminal, `${chosen.name} is converted`, subscriptions)
}
