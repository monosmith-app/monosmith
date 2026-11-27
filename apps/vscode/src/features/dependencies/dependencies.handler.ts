import * as vscode from 'vscode'
import type { CommandHost } from '../../shared/command-host/index.ts'
import { readFolderData, type ListNode } from '../../shared/list-view/index.ts'
import { openSmithTerminal } from '../../shared/smith-terminal/index.ts'
import { isWorkspaceTrusted } from '../../shared/workspace-trust/index.ts'
import { confirmRepair, confirmUpdate, tellNothingSafe, tellUpdateNeedsTrust } from './dependency-prompts.client.ts'
import { pickArguments, repairArguments, updateArguments } from './update-arguments.policy.ts'
import type { UpdatesReport } from './updates-report.contract.ts'
import { isRewritable } from './updates-tree.mapper.ts'

/**
 * Opens `smith deps up` in a terminal, in the folder the person picks: smith lists the updates there and the person picks them.
 * @param host - the folders and the Output channel of the extension
 */
export async function runUp (host: CommandHost): Promise<void> {
  const state = await host.pickReadyFolder()
  if (state?.location === undefined) return
  openSmithTerminal(state.location.executable, pickArguments(), state.root)
}

/**
 * Updates the packages of the Dependencies rows to their latest versions, in a terminal, after a
 * yes that shows the `smith deps up` command. With several rows selected (the terminal's marked
 * packages), they go in one command; a major release takes `--major`.
 * @param host - the folders and the Output channel of the extension
 * @param node - the row the command started on
 * @param selected - every selected row, when the person selected several
 */
export async function updateDependency (host: CommandHost, node: ListNode | undefined, selected?: readonly ListNode[]): Promise<void> {
  const rows = (selected !== undefined && selected.length > 1 ? selected : [node]).filter((row): row is ListNode => row?.context?.startsWith('dependency') === true)
  const tagged = rows.flatMap((row) => readFolderData(row.data) ?? [])
  const root = tagged[0]?.root
  const state = root === undefined ? undefined : host.registry.folders.get(root)
  const names = [...new Set(tagged.filter((item) => item.root === root).map((item) => item.data))]
  const major = rows.some((row) => row.context?.includes('major') === true)
  const args = updateArguments(names, major)
  if (args === undefined || state?.location === undefined) return
  if (!isWorkspaceTrusted()) {
    tellUpdateNeedsTrust()

    return
  }
  if (await confirmUpdate(names, `smith ${args.join(' ')}`)) openSmithTerminal(state.location.executable, args, state.root)
}

/**
 * Updates every patch and minor release of a folder in one `smith deps up` (the terminal's Update
 * all safe), in a terminal, after a yes that shows the command.
 * @param host - the folders and the Output channel of the extension
 * @param reports - what `updates` answered last, by folder
 */
export async function updateSafe (host: CommandHost, reports: ReadonlyMap<string, UpdatesReport>): Promise<void> {
  const state = await host.pickReadyFolder()
  if (state?.location === undefined) return
  const names = (reports.get(state.root)?.updates ?? []).filter((update) => update.level !== 'major' && isRewritable(update)).map((update) => update.name)
  const args = updateArguments(names)
  if (args === undefined) {
    tellNothingSafe()

    return
  }
  if (!isWorkspaceTrusted()) {
    tellUpdateNeedsTrust()

    return
  }
  if (await confirmUpdate(names, `smith ${args.join(' ')}`)) openSmithTerminal(state.location.executable, args, state.root)
}

/**
 * Opens `smith deps repair` in a terminal after a yes: smith asks there what to repair and shows the plan first.
 * @param host - the folders and the Output channel of the extension
 */
export async function repairDependencies (host: CommandHost): Promise<void> {
  const state = await host.pickReadyFolder()
  if (state?.location === undefined) return
  if (!isWorkspaceTrusted()) {
    tellUpdateNeedsTrust()

    return
  }
  const args = repairArguments()
  if (await confirmRepair(`smith ${args.join(' ')}`)) openSmithTerminal(state.location.executable, args, state.root)
}

/**
 * Opens the registry page of a package in the browser.
 * @param node - the Registry row of the package's detail
 */
export async function openRegistry (node: ListNode | undefined): Promise<void> {
  const url = readFolderData(node?.data)?.data
  if (url?.startsWith('https://') === true) await vscode.env.openExternal(vscode.Uri.parse(url))
}
