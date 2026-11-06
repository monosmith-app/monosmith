import * as vscode from 'vscode'
import type { FolderRegistry, FolderState } from '../folder-session/index.ts'
import { readFolderData, type ListNode } from '../list-view/index.ts'
import type { RefreshScope } from '../refresh/index.ts'
import type { SmithRun } from '../smith-run/index.ts'
import type { FolderWorkspace, WorkspaceSnapshot } from '../workspace/index.ts'
import { folderOfFile, toFolderPath } from '../workspace-folders/index.ts'
import { isWorkspaceTrusted } from '../workspace-trust/index.ts'
import { pickFolder, tellNeedsTrust, tellNoFolder } from './command-host-prompts.client.ts'
import type { CommandHost } from './command-host.contract.ts'
import { say } from '../words/index.ts'

/** What the host needs from the extension that owns it. */
export interface CommandHostDeps {
  registry:      FolderRegistry
  /** Runs smith in a folder with a time limit in milliseconds. */
  runSmith:      (executable: string, args: readonly string[], cwd: string, timeoutMs: number) => Promise<SmithRun>
  note:          (message: string) => void
  showOutput:    () => void
  focusOutput:   () => void
  fail:          (error: unknown) => void
  refreshFolder: (root: string, scope: RefreshScope, quiet: boolean) => Promise<void>
}

/**
 * Builds the host the commands of the editor use.
 * @param deps - the folders and the Output channel of the extension
 * @returns the host
 */
export function createCommandHost (deps: CommandHostDeps): CommandHost {
  const { registry, note } = deps
  const folderOf = async (): Promise<FolderState | undefined> => await pickFolder(registry.ready())
  const readyWorkspaces = (): FolderWorkspace[] => registry.ready()
    .filter((state) => state.snapshot !== undefined)
    .map((state) => ({ root: state.root, name: state.name, snapshot: state.snapshot as WorkspaceSnapshot }))

  return {
    registry,
    pickReadyFolder: folderOf,
    readyWorkspaces,
    fail:            deps.fail,
    focusOutput:     deps.focusOutput,
    fileTarget:      (argument) => {
      const tagged = readFolderData((argument as ListNode | undefined)?.data)
      if (tagged !== undefined) {
        const state = registry.folders.get(tagged.root)

        return state === undefined ? undefined : { state, file: tagged.data }
      }
      const uri = argument instanceof vscode.Uri ? argument : vscode.window.activeTextEditor?.document.uri
      if (uri?.scheme !== 'file') return
      const root = folderOfFile(registry.ready().map((state) => state.root), uri.fsPath)
      const state = root === undefined ? undefined : registry.folders.get(root)
      const file = root === undefined ? undefined : toFolderPath(root, uri.fsPath)

      return state === undefined || file === undefined ? undefined : { state, file }
    },
    folderFor: async (argument) => {
      const root = (argument as { root?: unknown } | undefined)?.root
      const given = typeof root === 'string' ? registry.folders.get(root) : undefined
      if (given?.session !== undefined) return given
      const state = await folderOf()
      if (state === undefined) tellNoFolder()

      return state
    },
    mayWrite: (what) => {
      if (isWorkspaceTrusted()) return true
      tellNeedsTrust(what)

      return false
    },
    smithIn: (state, log = true, timeoutMs = 60_000) => async (args) => {
      if (state.location === undefined) return { code: 1, stdout: '', stderr: say('vscode.command-host.smith-was-not-found-install-smith') }
      if (log) note(`$ smith ${args.join(' ')}${registry.states().length > 1 ? ` (${state.name})` : ''}`)

      return await deps.runSmith(state.location.executable, args, state.root, timeoutMs)
    },
    smithAt: (executable, root, name, timeoutMs) => async (args) => {
      note(`$ smith ${args.join(' ')} (${name})`)

      return await deps.runSmith(executable, args, root, timeoutMs)
    },
    note,
    showOutput:    deps.showOutput,
    refreshFolder: deps.refreshFolder,
    report:        (operation, state, outcome) => {
      note(say('vscode.command-host.monosmith-in', { operation, name: state.name, kind: outcome.kind, part: outcome.kind === 'refused' ? ` (${outcome.message ?? ''})` : '' }))
    },
  }
}
