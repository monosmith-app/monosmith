import type { CommandHost } from '../command-host/index.ts'
import { readFolderData, type ListNode } from '../list-view/index.ts'
import { sayDoctor, summarizeDoctor } from './doctor-report.mapper.ts'
import { fixDoctor } from './fix-doctor.use-case.ts'
import { confirmFix, pickFixAnswer, tellFixOutcome } from './fix-finding-prompts.client.ts'
import { fixFinding } from './fix-finding.use-case.ts'
import { confirmApplyFixes, tellChooseCheck, tellCheckPasses, tellDoctorSummary, tellFixesApplied, tellNoFix, tellNothingToCheck } from './health-prompts.client.ts'
import { runDoctor } from './run-doctor.use-case.ts'

/** What the Health view lets the commands do: read again, and forget what depends on it. */
export interface HealthView {
  /** The roots of the folders where the doctor proposes a fix. */
  fixableRoots: ReadonlySet<string>
  /** Reads the Health view from smith again. */
  reload:       () => Promise<void>
  /** Marks the Release Preview as out of date. */
  afterFix:     () => void
}

/**
 * Runs the doctor in every folder that runs smith and says what it found.
 * @param host - the folders and the Output channel of the extension
 */
export async function doctor (host: CommandHost): Promise<void> {
  const targets = host.registry.ready()
  if (targets.length === 0) {
    tellNothingToCheck()

    return
  }
  host.showOutput()
  for (const state of targets) {
    if (state.session === undefined) continue
    try {
      const summary = summarizeDoctor(await runDoctor(state.session.connection))
      host.note(`$ smith doctor${targets.length > 1 ? ` (${state.name})` : ''}`)
      for (const line of summary.lines) host.note(line)
      const message = sayDoctor(summary)
      host.note(message)
      tellDoctorSummary(message, summary.failed > 0 || summary.warnings > 0, host.focusOutput)
    } catch (error) {
      host.fail(error)
    }
  }
}

/**
 * Applies the fixes the doctor proposes, after smith shows what they change.
 * @param host - the folders and the Output channel of the extension
 * @param view - the Health view
 */
export async function fixHealth (host: CommandHost, view: HealthView): Promise<void> {
  const targets = host.registry.ready().filter((state) => view.fixableRoots.has(state.root))
  if (targets.length === 0) {
    tellNoFix()

    return
  }
  const several = host.registry.states().length > 1
  for (const state of targets) {
    if (state.session === undefined) continue
    const outcome = await fixDoctor(
      state.session.connection,
      async (preview) => await confirmApplyFixes(preview, several ? state.name : undefined),
      (line) => { host.note(`doctor | ${line}`) },
    )
    tellFixesApplied(outcome, several ? state.name : undefined)
  }
  await view.reload()
  view.afterFix()
}

/**
 * Fixes the check of a Health row: smith asks the fix's questions, shows what it changes, and applies it only on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param view - the Health view
 * @param argument - the row the command started on
 */
export async function fixFindingRow (host: CommandHost, view: HealthView, argument: unknown): Promise<void> {
  const tagged = readFolderData((argument as ListNode | undefined)?.data)
  const state = tagged === undefined ? undefined : host.registry.folders.get(tagged.root)
  if (tagged === undefined || state?.session === undefined) {
    tellChooseCheck()

    return
  }
  if (!host.mayWrite('a fix')) return
  const report = await runDoctor(state.session.connection)
  const finding = (report.findings ?? []).find((item) => item.check === tagged.data)
  if (finding === undefined) {
    tellCheckPasses()
    await view.reload()

    return
  }
  const outcome = await fixFinding({
    run:     host.smithIn(state, true, 600_000),
    pick:    pickFixAnswer,
    confirm: async (review) => await confirmFix(review, finding.fix?.description ?? finding.check),
  }, finding)
  tellFixOutcome(outcome)
  await view.reload()
  await host.refreshFolder(state.root, 'full', true)
}
