import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { isAbsolute, join } from 'node:path'
import * as vscode from 'vscode'
import { createAgentFlows, runSmith, writeAgentsFile, type AgentFilesState } from './agent-install/index.ts'
import { addDependencyHere } from './install-dependency/index.ts'
import { composeStackHere } from './compose-stack/index.ts'
import { addProjectHere } from './add-project/index.ts'
import { commitHere } from './commit-changes/index.ts'
import { projectAction } from './project-actions/index.ts'
import { addDeriveRows, createDeriveDiagnostics, deriveFile, readDeriveStatus, type DeriveReport, type DeriveView } from './derive-files/index.ts'
import { showWhy } from './why/index.ts'
import { bootstrapHere } from './bootstrap-launcher/index.ts'
import { installToolsHere } from './install-tools/index.ts'
import { tryHere } from './try-command/index.ts'
import { GUIDED_COMMANDS, guidedHere, readWebApps, showDivergences } from './guided-commands/index.ts'
import { newRecipeHere } from './new-recipe/index.ts'
import { swapHere } from './swap-choice/index.ts'
import { syncHere } from './sync-versions/index.ts'
import { undoHere } from './undo-change/index.ts'
import { openReport } from './workspace-report/index.ts'
import { adopt, findRepositoryOffers, type RepositoryOffer } from './adopt-offer/index.ts'
import { findBlankFolders, type BlankFolder } from './empty-folder/index.ts'
import { convertChoices } from './convert-workspace/index.ts'
import { createWorkspaceHere } from './new-workspace/index.ts'
import { setContextKey } from './context-keys/index.ts'
import { createCommandHost } from './command-host/index.ts'
import { connectFolder, createFolderRegistry, describeFolderNote, judgeViewState, type FolderState, type FoldersWithoutSmith } from './folder-session/index.ts'
import { doctor, fixFindingRow, fixHealth, readHealth, toHealthRows, type HealthReading, type HealthView } from './health/index.ts'
import { readUpdates, runUp, toDependencyRows, updateDependency } from './dependencies/index.ts'
import { createListView, loadAcross, type ListNode, type ListResult } from './list-view/index.ts'
import { createInsightDocuments, showInsight } from './project-insight/index.ts'
import { pickTask, pickTaskName, showStatusMenu } from './quick-pick/index.ts'
import { releaseHere } from './release-flow/index.ts'
import { explanationLines, readReleasePreview, toReleaseRows } from './release-preview/index.ts'
import { readSettings } from './settings/index.ts'
import { registerSmithTasks, type TaskFolder } from './task-provider/index.ts'
import {
  createTaskDiagnostics, createTaskLogDocuments, createTaskLogStore, logKey, registerOutputLinks, resolveSourcePath, sourceBases, LOG_SCHEME, type FolderProjects,
} from './task-output/index.ts'
import { createRefreshCoalescer, watchRefreshTriggers, type RefreshCoalescer, type RefreshScope } from './refresh/index.ts'
import { locateSmith, startSmith, type SmithExit } from './smith-client/index.ts'
import {
  describeOrigin, judgeSkew, readSmithVersion, createWarnedOnceStore,
} from './smith-version/index.ts'
import type { ExtensionStateApi } from './extension-api/index.ts'
import { createStatusBar } from './status-bar/index.ts'
import {
  cancelRun, createLastRunStore, createTaskStatusStore, describeEvent, describePlan, runTasks, toRunPlan,
  type RunEvent, type RunPlan, type RunTarget, type TaskStatusStore,
} from './task-run/index.ts'
import { readAffected, readWorkspace, type FolderWorkspace, type WorkspaceSnapshot } from './workspace/index.ts'
import { findSmithRoots, listFolders, onFoldersChanged, onManifestChanged, onTopLevelChanged } from './workspace-folders/index.ts'
import { askFilterText, createWorkspaceTree, type TaskFacts, type TreeNode, type ViewFilter, type WebAppEnvironments } from './workspace-view/index.ts'
import { isWorkspaceTrusted, onTrustGranted } from './workspace-trust/index.ts'

const INSTALL_COMMAND = 'npm i -D monosmith'
const UPDATE_COMMAND = 'npm i -D monosmith@latest'
const FILTER_KEY = 'monosmith.filter'

const registry = createFolderRegistry()
const { folders, states, ready } = registry
const agentFiles = new Map<string, AgentFilesState>()
/** Folders without a smith.toml that look like a repository smith can bring in. */
let offers: RepositoryOffer[] = []
/** Folders without a smith.toml and without a tool smith recognises: empty ones and ones with files. */
let blanks: BlankFolder[] = []

/** What the folders without a smith.toml hold, for the state of the view. */
function withoutSmith (): FoldersWithoutSmith {
  return {
    adoptable:    offers.length,
    empty:        blanks.filter((folder) => folder.kind === 'empty').length,
    unrecognised: blanks.filter((folder) => folder.kind === 'unrecognised').length,
  }
}

/** The names at the top of a folder, or undefined when it cannot be listed. Nothing is read. */
function listTopLevel (root: string): string[] | undefined {
  try {
    return readdirSync(root)
  } catch {
    return undefined
  }
}

/** A small text file, or undefined when it is missing, unreadable or too big to be a manifest. */
function readSmallText (path: string): string | undefined {
  try {
    return statSync(path).size > 1_000_000 ? undefined : readFileSync(path, 'utf8')
  } catch {
    return undefined
  }
}

function readyWorkspaces (): FolderWorkspace[] {
  return ready()
    .filter((state) => state.snapshot !== undefined)
    .map((state) => ({ root: state.root, name: state.name, snapshot: state.snapshot as WorkspaceSnapshot }))
}

function folderProjects (): FolderProjects[] {
  return ready().map((state) => ({
    root:     state.root,
    projects: state.snapshot?.projects.map((project) => ({ name: project.name, dir: project.dir })) ?? [],
  }))
}

function options (): { bail: boolean, noCache: boolean, parallel: number } {
  const { bail, noCache, parallel } = readSettings()

  return { bail, noCache, parallel }
}

function taskFolders (): TaskFolder[] {
  const found: TaskFolder[] = []
  for (const state of ready()) {
    if (state.location === undefined || state.snapshot === undefined) continue
    found.push({ root: state.root, executable: state.location.executable, tasks: state.snapshot.projects.flatMap((project) => project.tasks) })
  }

  return found
}

/** The file a path printed by a task means, found from the project the task belongs to. */
function resolveFile (root: string | undefined, task: string | undefined, file: string): string | undefined {
  return resolveSourcePath({ file, bases: sourceBases(folderProjects(), task, root), exists: existsSync })
}

const stopFolder = registry.stop

/**
 * Starts the extension: the MonoSmith view, its commands and the status bar,
 * one `smith serve` for each folder that has a `smith.toml`, all served by the
 * smith the team has installed.
 * @param context - the extension's lifetime
 */
export function activate (context: vscode.ExtensionContext): ExtensionStateApi {
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
  const logs = createTaskLogStore()
  const factsOf = (root: string, task: string): TaskFacts | undefined => {
    const result = statuses.get(root)?.resultOf(task)

    return result === undefined ? undefined : { ...result, hasLog: logs.has(logKey(root, task)) }
  }
  const stored = context.workspaceState.get<ViewFilter>(FILTER_KEY)
  const tree = createWorkspaceTree(factsOf, { text: stored?.text ?? '', affectedOnly: stored?.affectedOnly ?? false })
  const diagnostics = createTaskDiagnostics()
  const logDocs = createTaskLogDocuments(logs)
  const insight = createInsightDocuments()
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
    const viewState = judgeViewState(phases, withoutSmith())
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
    void setContextKey('monosmith.noProjects', readyStates.length > 0 && readyStates.every((state) => state.snapshot !== undefined && state.snapshot.projects.length === 0))
    void setContextKey('monosmith.repoKind', viewState === 'adoptable' ? offers[0]?.kind ?? '' : '')
  }

  const webApps = new Map<string, WebAppEnvironments[]>()

  const sync = (): void => {
    tree.show(states().map((state) => ({
      root:     state.root,
      name:     state.name,
      snapshot: state.phase === 'ready' ? state.snapshot : undefined,
      note:     describeFolderNote(state),
      ...(agentFiles.get(state.root) === 'missing' && { agentMissing: true }),
      ...(webApps.has(state.root) && { webApps: webApps.get(state.root) }),
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
        // The environments come from a second smith process: the view shows the projects first and the
        // Environments rows when they arrive.
        void readWebApps(host, state).then((found) => {
          if (stale()) return
          webApps.set(root, found)
          sync()
        })
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
    if (scope === 'full') {
      health.invalidate()
      release.invalidate()
      if (!quiet) dependencies.invalidate()
    } else {
      release.invalidate()
    }
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
    void checkAgentFiles(state, true)
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
    const coalescer = createRefreshCoalescer(readSettings().refreshDebounceMs, (scope) => { void refreshFolder(root, scope, true) })
    const watcher = watchRefreshTriggers(root, (scope) => {
      if (readSettings().refreshAuto) coalescer.request(scope)
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
      agentFiles.delete(root)
      unwatch(root)
    }
    offers = findRepositoryOffers(listFolders(), wanted, { exists: existsSync, read: readSmallText })
    blanks = findBlankFolders(listFolders(), new Set([...wanted, ...offers.map((offer) => offer.root)]), listTopLevel)
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

  const host = createCommandHost({
    registry,
    runSmith,
    note,
    showOutput:  () => { output.show(true) },
    focusOutput: () => { output.show() },
    fail,
    refreshFolder,
  })
  const { smithIn } = host
  const { checkAgentFiles, installFiles, talkToTheAgent } = createAgentFlows(host, { agentFiles, sync: () => { sync() }, workspaceState: context.workspaceState })
  const showPage = async (name: string, text: string): Promise<void> => { await insight.show(name, text) }

  /** Keeps what a task printed: its log for the row, and its failures for the Problems panel. */
  const keepOutput = (root: string, event: RunEvent, command: string): void => {
    const task = event.task
    if (task === undefined) return
    const key = logKey(root, task)
    switch (event.event) {
      case 'task.queued': {
        logs.begin(key, `$ ${command}`)
        diagnostics.begin(key)
        break
      }
      case 'task.output': {
        if (event.line === undefined) break
        logs.add(key, event.line)
        if (readSettings().problems) diagnostics.line(key, event.line, (file) => resolveFile(root, task, file))
        break
      }
      case 'task.finished': {
        const line = describeEvent(event)
        if (line !== undefined) logs.add(key, line)
        logDocs.changed(root, task)
        break
      }
      default: {
        break
      }
    }
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
      let command = 'smith run'
      const result = await runTasks(session.connection, { ...plan, runId }, (event) => {
        store.apply(event)
        const line = describeEvent(event)
        if (line !== undefined) output.appendLine(line)
        if (event.event === 'run.started') command = event.command ?? command
        keepOutput(root, event, command)
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

  const fixableRoots = new Set<string>()
  const deriveReports = new Map<string, DeriveReport>()
  const deriveProblems = createDeriveDiagnostics()

  /** Adds what `smith derive --check` finds to the rows of the doctor, and to the Problems panel. */
  const withDerivedFiles = async (state: FolderState, rows: ListResult): Promise<ListResult> => {
    const report = await readDeriveStatus(smithIn(state, false))
    deriveReports.set(state.root, report)
    if (readSettings().problems) deriveProblems.update(state.root, report)

    return addDeriveRows(rows, report)
  }

  const health = createListView({
    id:    'monosmith.health',
    title: 'Health',
    load:  async () => {
      fixableRoots.clear()
      deriveReports.clear()

      return await loadAcross(ready, 'smith is not running, so there is nothing to check.', async (state, connection) => {
        const reading: HealthReading = await readHealth(connection)
        if (reading.preview !== undefined) fixableRoots.add(state.root)

        return await withDerivedFiles(state, toHealthRows(reading.report, reading.preview))
      })
    },
    loaded: () => {
      void setContextKey('monosmith.fixable', fixableRoots.size > 0)
      void setContextKey('monosmith.derivedPaths', [...deriveReports].flatMap(([root, report]) => report.files.map((file) => vscode.Uri.file(join(root, file.path)).fsPath)))
    },
  })
  const healthView: HealthView = {
    fixableRoots,
    reload:   async () => { await health.reload() },
    afterFix: () => { release.invalidate() },
  }
  const deriveView: DeriveView = { reports: deriveReports, reload: healthView.reload }
  const dependencies = createListView({
    id:    'monosmith.dependencies',
    title: 'Dependencies',
    load:  async () => await loadAcross(ready, 'smith is not running, so dependencies cannot be read.', async (_state, connection) => toDependencyRows(await readUpdates(connection))),
  })
  const release = createListView({
    id:    'monosmith.release',
    title: 'Release Preview',
    load:  async () => await loadAcross(ready, 'smith is not running, so there is no release to preview.', async (state, connection) => {
      const preview = await readReleasePreview(connection)
      if ((preview.results ?? []).length > 0 || preview.error !== undefined) return toReleaseRows(preview)
      const said = await smithIn(state, false)(['release', '--dry-run'])

      return toReleaseRows(preview, said.code === 0 ? explanationLines(said.stdout) : [])
    }),
  })

  const setFilter = (next: ViewFilter): void => {
    tree.setFilter(next)
    void context.workspaceState.update(FILTER_KEY, next)
    void setContextKey('monosmith.affectedOnly', next.affectedOnly)
    void setContextKey('monosmith.filtered', next.text.trim() !== '')
  }

  const copyInstallCommand = async (): Promise<void> => {
    await vscode.env.clipboard.writeText(INSTALL_COMMAND)
    void vscode.window.showInformationMessage(`MonoSmith: copied "${INSTALL_COMMAND}". Run it in the workspace, then retry.`)
  }

  const command = <T = TreeNode>(id: string, action: (node?: T) => Promise<void>): vscode.Disposable => vscode.commands.registerCommand(id, async (node?: T) => {
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
    health,
    dependencies,
    release,
    diagnostics,
    deriveProblems,
    logDocs,
    insight,
    registerSmithTasks(taskFolders),
    registerOutputLinks({
      selector: [{ scheme: 'output' }, { scheme: LOG_SCHEME }],
      owns:     (document) => document.uri.scheme === LOG_SCHEME || document.uri.path.endsWith('MonoSmith'),
      resolve:  (document, task, reference) => {
        if (document.uri.scheme === LOG_SCHEME) {
          const which = logDocs.taskOf(document.uri)

          return resolveFile(which?.root, which?.task, reference.file)
        }
        // The output panel links absolute paths and paths relative to the folder itself; this adds the ones relative to a project.
        if (isAbsolute(reference.file)) return
        const viaProject = resolveFile(undefined, task, reference.file)
        const viaFolder = resolveSourcePath({ file: reference.file, bases: sourceBases(folderProjects(), undefined), exists: existsSync })

        return viaProject === undefined || viaProject === viaFolder ? undefined : viaProject
      },
    }),
    { dispose: () => { for (const root of watchers.keys()) unwatch(root) } },
    tree.onSelectionChange((node) => { void setContextKey('monosmith.selectionKind', node?.kind ?? '') }),
    onFoldersChanged(discover),
    onManifestChanged(discover),
    // A file appearing in a folder with no smith.toml turns an empty folder into one with files.
    onTopLevelChanged(() => { if (folders.size === 0) discover() }),
    onTrustGranted(() => { void restartAll() }),
    vscode.workspace.onDidChangeConfiguration((event) => {
      if (event.affectsConfiguration('monosmith.path')) void restartAll()
      else if (event.affectsConfiguration('monosmith.refresh')) {
        for (const root of watchers.keys()) {
          unwatch(root)
          watch(root)
        }
      }
    }),
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
    command('monosmith.doctor', async () => { await doctor(host) }),
    command('monosmith.showMenu', async () => {
      await showStatusMenu({ running: activeRun !== undefined, hasLastRun: lastRun.last() !== undefined, failed: failedTasks(), view: judgeViewState(states().map((state) => state.phase), withoutSmith()) })
    }),
    command('monosmith.showOutput', async () => { output.show() }),
    command('monosmith.openSettings', async () => { await vscode.commands.executeCommand('workbench.action.openSettings', 'monosmith') }),
    command('monosmith.copyInstallCommand', copyInstallCommand),
    command('monosmith.new', async () => { await createWorkspaceHere(host, blanks, refresh) }),
    command('monosmith.convert.choose', async () => { await convertChoices(host, offers, refresh) }),
    command('monosmith.adopt.start', async () => { await adopt(host, { offers, blanks }, context.subscriptions, 'start') }),
    command('monosmith.adopt.preview', async () => { await adopt(host, { offers, blanks }, context.subscriptions, 'preview') }),
    command<unknown>('monosmith.installAgent', async (argument) => { await installFiles('agent', argument) }),
    command<unknown>('monosmith.installMcp', async (argument) => { await installFiles('mcp', argument) }),
    command<unknown>('monosmith.talkToAgent', talkToTheAgent),
    command<unknown>('monosmith.agentFile', async (argument) => { await writeAgentsFile(host, argument) }),
    command<unknown>('monosmith.addProject', async (argument) => { await addProjectHere(host, argument) }),
    command<unknown>('monosmith.composeStack', async (argument) => { await composeStackHere(host, argument) }),
    command<unknown>('monosmith.addDependency', async (argument) => { await addDependencyHere(host, argument) }),
    command<unknown>('monosmith.addLibrary', async (argument) => { await addDependencyHere(host, argument, 'workspace') }),
    command<unknown>('monosmith.installPackage', async (argument) => { await addDependencyHere(host, argument, 'registry') }),
    command<unknown>('monosmith.commit', async (argument) => { await commitHere(host, argument) }),
    command<unknown>('monosmith.release', async (argument) => { await releaseHere(host, argument, async () => { await release.reload() }) }),
    command<unknown>('monosmith.removeProject', async (argument) => { await projectAction(host, argument, 'remove') }),
    command<unknown>('monosmith.changeKind', async (argument) => { await projectAction(host, argument, 'kind') }),
    command<unknown>('monosmith.unlinkLibrary', async (argument) => { await projectAction(host, argument, 'unlink') }),
    command<unknown>('monosmith.undo', async (argument) => { await undoHere(host, argument) }),
    command<unknown>('monosmith.syncVersions', async (argument) => { await syncHere(host, argument) }),
    command<unknown>('monosmith.swap', async (argument) => { await swapHere(host, argument) }),
    command<unknown>('monosmith.newRecipe', async (argument) => { await newRecipeHere(host, argument) }),
    command<unknown>('monosmith.divergences', async (argument) => { await showDivergences(host, showPage, argument) }),
    ...GUIDED_COMMANDS.map((spec) => command<unknown>(`monosmith.${spec.id}`, async (argument) => { await guidedHere(host, showPage, spec, argument) })),
    command<unknown>('monosmith.try', async (argument) => { await tryHere(host, showPage, argument) }),
    command<unknown>('monosmith.bootstrap', async (argument) => { await bootstrapHere(host, argument) }),
    command<unknown>('monosmith.installTools', async (argument) => { await installToolsHere(host, argument) }),
    command<unknown>('monosmith.openReport', async (argument) => { await openReport(host, argument) }),
    command<unknown>('monosmith.why', async (argument) => { await showWhy(host, insight, argument) }),
    command<unknown>('monosmith.derive.explain', async (argument) => { await deriveFile(host, insight, deriveView, argument, 'explain') }),
    command<unknown>('monosmith.derive.apply', async (argument) => { await deriveFile(host, insight, deriveView, argument, 'apply') }),
    command<unknown>('monosmith.derive.adopt', async (argument) => { await deriveFile(host, insight, deriveView, argument, 'adopt') }),
    command<unknown>('monosmith.derive.eject', async (argument) => { await deriveFile(host, insight, deriveView, argument, 'eject') }),
    command('monosmith.health.refresh', async () => { await health.reload() }),
    command('monosmith.health.fix', async () => { await fixHealth(host, healthView) }),
    command<unknown>('monosmith.health.fixFinding', async (argument) => { await fixFindingRow(host, healthView, argument) }),
    command('monosmith.dependencies.refresh', async () => { await dependencies.reload() }),
    command('monosmith.dependencies.up', async () => { await runUp(host) }),
    command<ListNode>('monosmith.dependencies.update', async (node) => { await updateDependency(host, node) }),
    command('monosmith.release.refresh', async () => { await release.reload() }),
    command('monosmith.explainProject', async (node) => { await showInsight(host, insight, node, 'explain') }),
    command('monosmith.showGraph', async (node) => { await showInsight(host, insight, node, 'graph') }),
    command('monosmith.showTaskOutput', async (node) => {
      const target = node ?? tree.selected()
      if (target?.kind !== 'task' || target.project === undefined || target.task === undefined) return
      const task = `${target.project}:${target.task}`
      if (logs.has(logKey(target.root, task))) await logDocs.open(target.root, task)
      else void vscode.window.showInformationMessage(`MonoSmith: ${task} has not run since the editor started, so there is no output to show.`)
    }),
    command('monosmith.clearProblems', async () => {
      diagnostics.clearAll()
      deriveProblems.clearAll()
    }),
    command('monosmith.affectedOnly.on', async () => { setFilter({ ...tree.filter(), affectedOnly: true }) }),
    command('monosmith.affectedOnly.off', async () => { setFilter({ ...tree.filter(), affectedOnly: false }) }),
    command('monosmith.filter', async () => { askFilterText(tree.filter().text, (text) => { setFilter({ ...tree.filter(), text }) }) }),
    command('monosmith.clearFilter', async () => { setFilter({ text: '', affectedOnly: tree.filter().affectedOnly }) }),
    command('monosmith.openWalkthrough', async () => {
      await vscode.commands.executeCommand('workbench.action.openWalkthrough', 'monosmith.monosmith-vscode#monosmith.start', false)
    }),
  )
  setFilter(tree.filter())
  void setContextKey('monosmith.running', false)
  void setContextKey('monosmith.hasLastRun', false)
  void setContextKey('monosmith.selectionKind', '')
  discover()

  return {
    viewState:  () => judgeViewState(states().map((state) => state.phase), withoutSmith()),
    projects:   () => readyWorkspaces().flatMap((folder) => folder.snapshot.projects.map((project) => project.name)),
    taskStatus: (task) => {
      for (const store of statuses.values()) {
        const status = store.statusOf(task)
        if (status !== undefined) return status
      }

      return
    },
    isRunning:  () => activeRun !== undefined,
    roots:      () => ready().map((state) => state.root),
    agentFiles: (root) => agentFiles.get(root),
    offers:     () => offers.map((offer) => offer.kind),
  }
}

/** Ends every `smith serve` with the extension. */
export function deactivate (): void {
  for (const state of folders.values()) state.session?.stop()
  folders.clear()
}
