import { basename } from 'node:path'
import type { CommandHost } from '../command-host/index.ts'
import type { InsightDocuments } from '../project-insight/index.ts'
import { applyActionFor } from './derive-arguments.mapper.ts'
import { toExplainDocument, toTextDocument } from './derive-explain.mapper.ts'
import { confirmDeriveChange, tellChooseDerivedFile, tellDeriveChange, tellNothingToWrite } from './derive-prompts.client.ts'
import type { DeriveAction, DeriveReport } from './derive-report.contract.ts'
import { changeDerivedFile, explainDerivedFile, previewDerive } from './derive-files.use-case.ts'
import { say } from '../words/index.ts'

/** What the Health view lets the derive commands do. */
export interface DeriveView {
  /** What `smith derive --check` found in each folder, by root. */
  reports: ReadonlyMap<string, DeriveReport>
  /** Reads the Health view from smith again. */
  reload:  () => Promise<void>
}

/**
 * Explains, applies, adopts or ejects a derived file: smith shows what it changes, and nothing changes before a yes.
 * @param host - the folders and the Output channel of the extension
 * @param insight - where the pages are shown
 * @param view - the Health view
 * @param argument - what the editor passed to the command
 * @param action - what to do with the file
 */
export async function deriveFile (host: CommandHost, insight: InsightDocuments, view: DeriveView, argument: unknown, action: 'explain' | 'apply' | 'adopt' | 'eject'): Promise<void> {
  const target = host.fileTarget(argument)
  if (target === undefined) {
    tellChooseDerivedFile()

    return
  }
  const { state, file } = target
  const run = host.smithIn(state)
  const explain = async (): Promise<void> => { await insight.show(`${basename(file)}-derive.md`, toExplainDocument(file, await explainDerivedFile(run, file))) }
  if (action === 'explain') {
    await explain()

    return
  }
  if (!host.mayWrite(say('vscode.derive-files.changing-a-derived-file'))) return
  const status = view.reports.get(state.root)?.files.find((item) => item.path === file)?.status
  if (action === 'apply' && status === undefined) {
    tellNothingToWrite(file)

    return
  }
  const change: DeriveAction & ('apply' | 'replace' | 'adopt' | 'eject') = action === 'apply' && status !== undefined ? applyActionFor(status) : action
  let commit: boolean
  for (;;) {
    const answer = await confirmDeriveChange(change, change === 'apply' ? undefined : file)
    if (answer === 'no') return
    if (answer === 'yes' || answer === 'yes-commit') {
      commit = answer === 'yes-commit'
      break
    }
    if (answer === 'preview') await insight.show('derive-preview.md', toTextDocument(say('vscode.derive-files.derived-files-what-smith-derive-would'), await previewDerive(run)))
    else await explain()
  }
  const result = await changeDerivedFile(run, change, change === 'apply' ? undefined : file, commit)
  const printed = result.ok ? result.output.split(/\r?\n/).filter(Boolean) : []
  for (const line of printed) host.note(line)
  tellDeriveChange(change, change === 'apply' ? undefined : file, result.ok, result.ok ? '' : result.message)
  await view.reload()
  await host.refreshFolder(state.root, 'full', true)
}
