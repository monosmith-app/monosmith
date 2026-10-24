import { existsSync } from 'node:fs'
import * as vscode from 'vscode'
import { setContextKey } from './context-keys/index.ts'
import { connectFolder, describeFolderNote, judgeViewState, type FolderState } from './folder-session/index.ts'
import { runDoctor, sayDoctor, summarizeDoctor } from './health/index.ts'
import { pickTask, pickTaskName, showStatusMenu } from './quick-pick/index.ts'
import { createRefreshCoalescer, watchRefreshTriggers, type RefreshCoalescer, type RefreshScope } from './refresh/index.ts'
import { locateSmith, startSmith, type SmithExit } from './smith-client/index.ts'
import {
  describeOrigin, judgeSkew, readSmithVersion, createWarnedOnceStore,
} from './smith-version/index.ts'
import { createStatusBar } from './status-bar/index.ts'
import {
  cancelRun, createLastRunStore, createTaskStatusStore, describeEvent, describePlan, runTasks, toRunPlan,
  type RunPlan, type RunTarget, type TaskStatusStore,
} from './task-run/index.ts'
import { readAffected, readWorkspace, type FolderWorkspace, type WorkspaceSnapshot } from './workspace/index.ts'
import { findSmithRoots, listFolders, onFoldersChanged } from './workspace-folders/index.ts'
import { createWorkspaceTree, type TreeNode } from './workspace-view/index.ts'
import { isWorkspaceTrusted, onTrustGranted } from './workspace-trust/index.ts'

const INSTALL_COMMAND = 'npm i -D monosmith'
const UPDATE_COMMAND = 'npm i -D monosmith@latest'
const REFRESH_DELAY_MS = 1500

const folders = new Map<string, FolderState>()

function states (): FolderState[] {
  return Array.from(folders, ([, state]) => state)
}

function readyWorkspaces (): FolderWorkspace[] {
  return ready()
    .filter((state) => state.snapshot !== undefined)
    .map((state) => ({ root: state.root, name: state.name, snapshot: state.snapshot as WorkspaceSnapshot }))
}

function ready (): FolderState[] {
  return states().filter((state) => state.phase === 'ready' && state.session !== undefined)
}

function stopFolder (root: string): void {
  const state = folders.get(root)
  state?.session?.stop()
  if (state !== undefined) state.session = undefined
}

/**
 * Starts the extension: the MonoSmith view, its commands and the status bar,
 * one `smith serve` for each folder that has a `smith.toml`, all served by the
 * smith the team has installed.
 * @param context - the extension's lifetime
 */
export function activate (context: vscode.ExtensionContext): void {
  const output = vscode.window.createOutputChannel('MonoSmith')
  const statuses = new Map<string, TaskStatusStore>()
  const watchers = new Map<string, { coalescer: RefreshCoalescer, watcher: vscode.Disposable }>()
  const generations = new Map<string, number>()
  const connecting = new Set<string>()
  const terminalVersions = new Map<string, string | undefined>()
  const warned = createWarnedOnceStore()
  const lastRun = createLastRunStore()
  const statusOf = (root: string): TaskStatusStore => {
    let store = statuses.get(root)
    if (store === undefined) {
      store = createTaskStatusStore()
      store.onChange(() => {
        tree.redraw()
        paint()
      })
      statuses.set(root, store)
    }

    return store
  }
  const tree = createWorkspaceTree((root, task) => statuses.get(root)?.statusOf(task))
  const bar = createStatusBar('monosmith.showMenu')
  let activeRun: { id: string, root: string } | undefined

  const failedTasks = (): number => {
    let failed = 0
    for (const store of statuses.values()) failed += store.failedCount()

    return failed
  }
  const note = (message: string): void => { output.appendLine(message) }
  const fail = (error: unknown): void => {
    const message = error instanceof Error ? error.message : String(error)
    note(`MonoSmith: ${message}`)
    void vscode.window.showErrorMessage(`MonoSmith: ${message}`, 'Show output').then((choice) => { if (choice === 'Show output') output.show() })
  }

  const paint = (): void => {
    const phases = states().map((state) => state.phase)
    const viewState = judgeViewState(phases)
    const readyStates = ready()
    const many = states().length > 1
    const origin = readyStates.flatMap((state) => {
      if (state.location === undefined || state.session === undefined) return []
      const lines = describeOrigin({
        location:         state.location,
        version:          state.session.version,
        skippedUntrusted: state.skippedUntrusted,
        terminalVersion:  state.pathSmith === undefined ? undefined : terminalVersions.get(state.pathSmith),
      })

      return many ? [`${state.name}:`, ...lines] : lines
    })
    const affected = readyStates.some((state) => state.snapshot !== undefined)
      ? readyStates.reduce((sum, state) => sum + (state.snapshot?.affected.size ?? 0), 0)
      : undefined
    bar.update({
      state:   viewState,
      version: readyStates[0]?.session?.version,
      affected,
      running: activeRun !== undefined,
      failed:  failedTasks(),
      origin,
    })
    void setContextKey('monosmith.state', viewState)
  }

  const sync = (): void => {
    tree.show(states().map((state) => ({
      root:     state.root,
      name:     state.name,
      snapshot: state.phase === 'ready' ? state.snapshot : undefined,
      note:     describeFolderNote(state),
    })))
    paint()
  }

  const warnAboutSkew = async (state: FolderState): Promise<void> => {
    if (state.location === undefined || state.session === undefined || state.pathSmith === undefined || (state.location.executable === state.pathSmith)) return
    if (!terminalVersions.has(state.pathSmith)) terminalVersions.set(state.pathSmith, await readSmithVersion(state.pathSmith))
    paint()
    const warning = judgeSkew({
      editorSmith:     state.location.executable,
      editorVersion:   state.session.version,
      editorSource:    state.location.source,
      terminalSmith:   state.pathSmith,
      terminalVersion: terminalVersions.get(state.pathSmith),
    })
    if (warning === undefined || !warned.firstTime('skew')) return
    note(warning)
    const copy = `Copy "${UPDATE_COMMAND}"`
    const choice = await vscode.window.showWarningMessage(warning, 'Open settings', copy)
    if (choice === 'Open settings') await vscode.commands.executeCommand('monosmith.openSettings')
    else if (choice === copy) await vscode.env.clipboard.writeText(UPDATE_COMMAND)
  }

  const refreshFolder = async (root: string, scope: RefreshScope, quiet: boolean): Promise<void> => {
    const state = folders.get(root)
    const session = state?.session
    if (state === undefined || session === undefined || state.phase !== 'ready') return
    const token = (generations.get(root) ?? 0) + 1
    generations.set(root, token)
    const stale = (): boolean => generations.get(root) !== token || folders.get(root)?.session !== session
    try {
      if (scope === 'full' || state.snapshot === undefined) {
        const snapshot = await readWorkspace(session.connection)
        if (stale()) return
        state.snapshot = snapshot
      } else {
        const affected = await readAffected(session.connection)
        if (stale()) return
        state.snapshot = { ...state.snapshot, affected }
      }
    } catch (error) {
      if (stale()) return
      if (quiet) note(`MonoSmith: could not refresh ${state.name}: ${error instanceof Error ? error.message : String(error)}`)
      else fail(error)
    }
    sync()
  }

  const connect = async (root: string): Promise<void> => {
    const state = folders.get(root)
    if (state === undefined || connecting.has(root)) return
    connecting.add(root)
    stopFolder(root)
    Object.assign(state, { phase: 'starting', error: undefined, snapshot: undefined })
    sync()
    const trusted = isWorkspaceTrusted()
    const outcome = await connectFolder({
      locate: (folder) => locateSmith({
        setting:       vscode.workspace.getConfiguration('monosmith').get<string>('path', ''),
        workspaceRoot: folder,
        trusted,
        platform:      process.platform,
        pathVariable:  process.env.PATH ?? '',
        exists:        existsSync,
      }),
      start: async (executable, folder) => await startSmith(executable, folder, {
        onStderr: (line) => { note(`[${state.name}] ${line}`) },
        onExit:   (exit) => { handleExit(root, exit) },
      }),
    }, root)
    connecting.delete(root)
    if (folders.get(root) !== state) {
      outcome.session?.stop()

      return
    }
    Object.assign(state, outcome)
    if (outcome.phase === 'failed') note(`MonoSmith: smith could not start in ${state.name}: ${outcome.error ?? ''}`)
    if (outcome.session?.protocol === 'newer') note('smith speaks a newer protocol than this extension knows; it keeps working with what it knows.')
    sync()
    if (outcome.phase !== 'ready') {
      return
    }

    await refreshFolder(root, 'full', false)
    void warnAboutSkew(state)
  }

  const handleExit = (root: string, exit: SmithExit): void => {
    const state = folders.get(root)
    if (state === undefined) return
    const why = exit.signal === null ? `exit code ${String(exit.code)}` : `signal ${exit.signal}`
    Object.assign(state, { phase: 'failed', session: undefined, error: `smith serve ended (${why})${exit.stderr === '' ? '' : `: ${exit.stderr.split('\n').at(-1) ?? ''}`}` })
    note(`MonoSmith: smith serve ended in ${state.name} (${why}).`)
    if (activeRun?.root === root) {
      activeRun = undefined
      void setContextKey('monosmith.running', false)
    }
    sync()
    void vscode.window.showErrorMessage(`MonoSmith: smith stopped unexpectedly in ${state.name} (${why}).`, 'Restart', 'Show output').then((choice) => {
      if (choice === 'Restart') void connect(root)
      else if (choice === 'Show output') output.show()
    })
  }

  const watch = (root: string): void => {
    const coalescer = createRefreshCoalescer(REFRESH_DELAY_MS, (scope) => { void refreshFolder(root, scope, true) })
    const watcher = watchRefreshTriggers(root, (scope) => {
      coalescer.request(scope)
    })
    watchers.set(root, { coalescer, watcher })
  }

  const unwatch = (root: string): void => {
    const entry = watchers.get(root)
    entry?.coalescer.cancel()
    entry?.watcher.dispose()
    watchers.delete(root)
  }

  const discover = (): void => {
    const roots = findSmithRoots(listFolders(), existsSync)
    const wanted = new Set(roots.map((folder) => folder.root))
    for (const root of folders.keys()) {
      if (wanted.has(root)) continue
      stopFolder(root)
      folders.delete(root)
      statuses.delete(root)
      unwatch(root)
    }
    const added = roots.filter((folder) => !folders.has(folder.root))
    for (const folder of added) {
      folders.set(folder.root, { root: folder.root, name: folder.name, phase: 'starting', skippedUntrusted: false })
      watch(folder.root)
    }
    sync()
    for (const folder of added) void connect(folder.root)
  }

  const restartAll = async (): Promise<void> => {
    await Promise.all(states().map(async (state) => { await connect(state.root) }))
  }

  const refresh = async (): Promise<void> => {
    await tree.withProgress(async () => {
      await Promise.all(states().map(async (state) => {
        if (state.phase === 'ready') await refreshFolder(state.root, 'full', false)
        else await connect(state.root)
      }))
    })
  }

  const run = async (root: string, plan: RunPlan | undefined, confirm = false): Promise<void> => {
    if (plan === undefined) {
      void vscode.window.showInformationMessage('MonoSmith: there is nothing to run there.')

      return
    }
    if (!isWorkspaceTrusted()) {
      const choice = await vscode.window.showWarningMessage('MonoSmith: running tasks needs a trusted workspace; they execute code from it.', 'Manage Workspace Trust')
      if (choice !== undefined) await vscode.commands.executeCommand('workbench.trust.manage')

      return
    }
    if (activeRun !== undefined) {
      void vscode.window.showWarningMessage('MonoSmith: a run is already in progress.')

      return
    }
    const label = describePlan(plan)
    if (confirm) {
      const choice = await vscode.window.showWarningMessage(`MonoSmith: run ${label}?`, { modal: true }, 'Run')
      if (choice !== 'Run') return
    }
    const state = folders.get(root)
    if (state === undefined) return
    if (state.session === undefined) await connect(root)
    const session = state.session
    if (session === undefined) return
    lastRun.remember({ root, plan })
    void setContextKey('monosmith.hasLastRun', true)
    const runId = `ide-${Date.now()}`
    activeRun = { id: runId, root }
    void setContextKey('monosmith.running', true)
    paint()
    output.show(true)
    const store = statusOf(root)
    try {
      const result = await runTasks(session.connection, { ...plan, runId }, (event) => {
        store.apply(event)
        const line = describeEvent(event)
        if (line !== undefined) output.appendLine(line)
      })
      if (!result.ok) void vscode.window.showWarningMessage(`MonoSmith: ${label} did not pass. See the output.`, 'Show output').then((choice) => { if (choice === 'Show output') output.show() })
    } catch (error) {
      note(`MonoSmith: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      activeRun = undefined
      void setContextKey('monosmith.running', false)
      paint()
      void refreshFolder(root, 'affected', true)
    }
  }

  const options = (): { bail: boolean, noCache: boolean } => {
    const config = vscode.workspace.getConfiguration('monosmith.run')

    return { bail: config.get<boolean>('bail', false), noCache: config.get<boolean>('noCache', false) }
  }

  const runTarget = async (root: string, target: RunTarget, confirm = false): Promise<void> => {
    const projects = folders.get(root)?.snapshot?.projects ?? []
    await run(root, toRunPlan(target, projects, options()), confirm)
  }

  const runNode = async (node: TreeNode | undefined): Promise<void> => {
    if (node?.kind === 'task' && node.project !== undefined && node.task !== undefined) {
      await runTarget(node.root, { kind: 'task', project: node.project, task: node.task })
    } else if (node === undefined) {
      const picked = await pickTask(readyWorkspaces())
      if (picked !== undefined) await runTarget(picked.root, picked.target)
    } else {
      await runTarget(node.root, { kind: 'projects', projects: node.projects }, node.projects.length > 1)
    }
  }

  const runAcross = async (scope: 'affected' | 'all'): Promise<void> => {
    const picked = await pickTaskName(readyWorkspaces(), scope)
    if (picked !== undefined) await runTarget(picked.root, picked.target, scope === 'all')
  }

  const cancel = async (): Promise<void> => {
    const current = activeRun
    const session = current === undefined ? undefined : folders.get(current.root)?.session
    if (current === undefined || session === undefined) return
    await cancelRun(session.connection, current.id)
  }

  const doctor = async (): Promise<void> => {
    const targets = ready()
    if (targets.length === 0) {
      void vscode.window.showInformationMessage('MonoSmith: smith is not running, so there is nothing to check.')

      return
    }
    output.show(true)
    for (const state of targets) {
      if (state.session === undefined) continue
      try {
        const summary = summarizeDoctor(await runDoctor(state.session.connection))
        output.appendLine(`$ smith doctor${targets.length > 1 ? ` (${state.name})` : ''}`)
        for (const line of summary.lines) output.appendLine(line)
        const message = sayDoctor(summary)
        output.appendLine(message)
        const show = summary.failed > 0 || summary.warnings > 0 ? vscode.window.showWarningMessage : vscode.window.showInformationMessage
        void show(`MonoSmith: ${message}`, 'Show output').then((choice) => { if (choice === 'Show output') output.show() })
      } catch (error) {
        fail(error)
      }
    }
  }

  const copyInstallCommand = async (): Promise<void> => {
    await vscode.env.clipboard.writeText(INSTALL_COMMAND)
    void vscode.window.showInformationMessage(`MonoSmith: copied "${INSTALL_COMMAND}". Run it in the workspace, then retry.`)
  }

  const command = (id: string, action: (node?: TreeNode) => Promise<void>): vscode.Disposable => vscode.commands.registerCommand(id, async (node?: TreeNode) => {
    try {
      await action(node)
    } catch (error) {
      fail(error)
    }
  })

  context.subscriptions.push(
    output,
    bar,
    tree,
    { dispose: () => { for (const root of watchers.keys()) unwatch(root) } },
    tree.onSelectionChange((node) => { void setContextKey('monosmith.selectionKind', node?.kind ?? '') }),
    onFoldersChanged(discover),
    onTrustGranted(() => { void restartAll() }),
    vscode.workspace.onDidChangeConfiguration((event) => { if (event.affectsConfiguration('monosmith.path')) void restartAll() }),
    command('monosmith.refresh', refresh),
    command('monosmith.restart', restartAll),
    command('monosmith.runTask', runNode),
    command('monosmith.runSelected', async () => {
      // Arrow keys move the focus, not the selection; Enter selects the focused row first.
      await vscode.commands.executeCommand('list.select')
      const node = tree.selected()
      if (node?.kind === 'task') await runNode(node)
    }),
    command('monosmith.runAllTasks', async (node) => { await runNode(node ?? tree.selected()) }),
    command('monosmith.runAffected', async () => { await runAcross('affected') }),
    command('monosmith.runAll', async () => { await runAcross('all') }),
    command('monosmith.rerunLast', async () => {
      const last = lastRun.last()
      if (last === undefined) void vscode.window.showInformationMessage('MonoSmith: nothing has run yet.')
      else await run(last.root, last.plan)
    }),
    command('monosmith.cancelRun', cancel),
    command('monosmith.doctor', doctor),
    command('monosmith.showMenu', async () => {
      await showStatusMenu({ running: activeRun !== undefined, hasLastRun: lastRun.last() !== undefined, failed: failedTasks() })
    }),
    command('monosmith.showOutput', async () => { output.show() }),
    command('monosmith.openSettings', async () => { await vscode.commands.executeCommand('workbench.action.openSettings', 'monosmith') }),
    command('monosmith.copyInstallCommand', copyInstallCommand),
  )
  void setContextKey('monosmith.running', false)
  void setContextKey('monosmith.hasLastRun', false)
  void setContextKey('monosmith.selectionKind', '')
  discover()
}

/** Ends every `smith serve` with the extension. */
export function deactivate (): void {
  for (const state of folders.values()) state.session?.stop()
  folders.clear()
}
