import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { basename, isAbsolute, join } from 'node:path'
import * as vscode from 'vscode'
import {
  confirmAgentFile, confirmInstall, explainNoClaudeCode, hasClaudeCode, installAgent, judgeAgentFiles, offerAgent, optOutOfAgent, runSmith, shouldOfferAgent, talkToAgent, tellAgentFileOutcome, tellInstallOutcome, writeAgentFile,
  type AgentFilesState, type InstallScope,
} from './agent-install/index.ts'
import { addProject, askAddRequest, confirmAddPlan, parseChoices, tellAddOutcome } from './add-project/index.ts'
import {
  addDeriveRows, applyActionFor, changeDerivedFile, confirmDeriveChange, createDeriveDiagnostics, explainDerivedFile, previewDerive, readDeriveStatus, toTextDocument, tellDeriveChange, toExplainDocument,
  type DeriveAction, type DeriveReport,
} from './derive-files/index.ts'
import { readFileWhy, readSettingWhy, readTaskWhy } from './why/index.ts'
import { confirmLauncher, tellLauncherOutcome, writeLauncher } from './bootstrap-launcher/index.ts'
import { askTryCommand, offerTryApply, tellTryOutcome, tryCommand } from './try-command/index.ts'
import { describeFailure } from './smith-run/index.ts'
import { askSwapRequest, confirmSwapPlan, swapChoice, tellSwapOutcome } from './swap-choice/index.ts'
import { confirmUndo, tellUndoOutcome, undoLastChange } from './undo-change/index.ts'
import { reportFile, writeReport } from './workspace-report/index.ts'
import { askWhereToMove, toAdoptArguments, toMoveArguments, findRepositoryOffers, type AdoptAction, type RepositoryOffer } from './adopt-offer/index.ts'
import { findBlankFolders, toNewArguments, type BlankFolder } from './empty-folder/index.ts'
import { setContextKey } from './context-keys/index.ts'
import { connectFolder, describeFolderNote, judgeViewState, type FolderState, type FoldersWithoutSmith } from './folder-session/index.ts'
import { fixDoctor, readHealth, runDoctor, sayDoctor, summarizeDoctor, toHealthRows, type HealthReading } from './health/index.ts'
import { pickArguments, readUpdates, toDependencyRows, updateArguments } from './dependencies/index.ts'
import { createListView, groupByFolder, readFolderData, tagWithFolder, type ListNode, type ListResult } from './list-view/index.ts'
import { createInsightDocuments, readExplain, readGraph } from './project-insight/index.ts'
import { pickProject, pickTask, pickTaskName, showStatusMenu } from './quick-pick/index.ts'
import { readReleasePreview, toReleaseRows } from './release-preview/index.ts'
import { readSettings } from './settings/index.ts'
import { openSmithTerminal } from './smith-terminal/index.ts'
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
import { findSmithRoots, folderOfFile, listFolders, onFoldersChanged, onManifestChanged, onTopLevelChanged, toFolderPath } from './workspace-folders/index.ts'
import { askFilterText, createWorkspaceTree, type TaskFacts, type TreeNode, type ViewFilter } from './workspace-view/index.ts'
import { isWorkspaceTrusted, onTrustGranted } from './workspace-trust/index.ts'

const INSTALL_COMMAND = 'npm i -D monosmith'
const UPDATE_COMMAND = 'npm i -D monosmith@latest'
const FILTER_KEY = 'monosmith.filter'
const AGENT_ASKED_KEY = 'monosmith.agentOffered'
const AGENT_FILE = join('.claude', 'agents', 'smith.md')

const folders = new Map<string, FolderState>()
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
const INSTALL_GLOBAL_COMMAND = 'npm i -g monosmith'

/** A small text file, or undefined when it is missing, unreadable or too big to be a manifest. */
function readSmallText (path: string): string | undefined {
  try {
    return statSync(path).size > 1_000_000 ? undefined : readFileSync(path, 'utf8')
  } catch {
    return undefined
  }
}

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

/** Reads a list view's rows from every folder that runs smith, and lays them out by folder. */
async function loadAcross (nothing: string, read: (state: FolderState, connection: NonNullable<FolderState['session']>['connection']) => Promise<ListResult>): Promise<ListResult> {
  const targets = ready()
  if (targets.length === 0) return { nodes: [], message: nothing }
  const results = await Promise.all(targets.map(async (state) => {
    try {
      return { state, result: await read(state, (state.session as NonNullable<FolderState['session']>).connection) }
    } catch (error) {
      return { state, result: { nodes: [], message: `${state.name}: ${error instanceof Error ? error.message : String(error)}` } satisfies ListResult }
    }
  }))
  const nodes = groupByFolder(results.map(({ state, result }) => ({ root: state.root, name: state.name, nodes: tagWithFolder(result.nodes, state.root) })))
  const messages = results.flatMap(({ result }) => result.message ?? [])
  const summaries = results.flatMap(({ result }) => result.summary ?? [])

  return {
    nodes,
    ...(messages.length > 0 && (nodes.length === 0 || results.length === 1) && { message: messages.join(' ') }),
    ...(summaries.length > 0 && { summary: results.length === 1 ? summaries[0] : summaries.join(' · ') }),
  }
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
    void setContextKey('monosmith.repoKind', viewState === 'adoptable' ? offers[0]?.kind ?? '' : '')
  }

  const sync = (): void => {
    tree.show(states().map((state) => ({
      root:     state.root,
      name:     state.name,
      snapshot: state.phase === 'ready' ? state.snapshot : undefined,
      note:     describeFolderNote(state),
      ...(agentFiles.get(state.root) === 'missing' && { agentMissing: true }),
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

  const fixableRoots = new Set<string>()
  const deriveReports = new Map<string, DeriveReport>()
  const deriveProblems = createDeriveDiagnostics()

  /** Runs smith in the folder of a state; the caller checked that the state has a smith. */
  const smithIn = (state: FolderState, log = true, timeoutMs = 60_000): ((args: string[]) => ReturnType<typeof runSmith>) => async (args) => {
    if (state.location === undefined) return { code: 1, stdout: '', stderr: '✖ smith was not found\n  → install smith in the workspace, or set monosmith.path\n' }
    if (log) note(`$ smith ${args.join(' ')}${states().length > 1 ? ` (${state.name})` : ''}`)

    return await runSmith(state.location.executable, args, state.root, timeoutMs)
  }

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

      return await loadAcross('smith is not running, so there is nothing to check.', async (state, connection) => {
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
  const dependencies = createListView({
    id:    'monosmith.dependencies',
    title: 'Dependencies',
    load:  async () => await loadAcross('smith is not running, so dependencies cannot be read.', async (_state, connection) => toDependencyRows(await readUpdates(connection))),
  })
  const release = createListView({
    id:    'monosmith.release',
    title: 'Release Preview',
    load:  async () => await loadAcross('smith is not running, so there is no release to preview.', async (_state, connection) => toReleaseRows(await readReleasePreview(connection))),
  })

  const fixHealth = async (): Promise<void> => {
    const targets = ready().filter((state) => fixableRoots.has(state.root))
    if (targets.length === 0) {
      void vscode.window.showInformationMessage('MonoSmith: doctor proposes no fix. The remedies listed under a check say what to do by hand.')

      return
    }
    for (const state of targets) {
      if (state.session === undefined) continue
      const outcome = await fixDoctor(state.session.connection, async (preview) => {
        const choice = await vscode.window.showWarningMessage(
          `MonoSmith: apply ${preview.titles.length} ${preview.titles.length === 1 ? 'fix' : 'fixes'}${states().length > 1 ? ` in ${state.name}` : ''}?`,
          { modal: true, detail: `${preview.text}\n\nThe files are changed as shown. Review them with git afterwards.` },
          'Apply fixes',
        )

        return choice === 'Apply fixes'
      }, (line) => { output.appendLine(`doctor | ${line}`) })
      if (outcome === 'fixed') void vscode.window.showInformationMessage(`MonoSmith: the fixes were applied${states().length > 1 ? ` in ${state.name}` : ''}.`)
      else if (outcome === 'incomplete') void vscode.window.showInformationMessage('MonoSmith: the fixes were applied. Some checks still fail and need their remedy by hand; the Health view lists them.')
    }
    await health.reload()
    release.invalidate()
  }

  const folderOf = async (): Promise<FolderState | undefined> => {
    const targets = ready()
    if (targets.length <= 1) return targets[0]
    const picked = await vscode.window.showQuickPick(targets.map((state) => ({ label: state.name, description: state.root, state })), { title: 'Choose a folder', placeHolder: 'Folder' })

    return picked?.state
  }

  const runUp = async (): Promise<void> => {
    const state = await folderOf()
    if (state?.location === undefined) return
    openSmithTerminal(state.location.executable, pickArguments(), state.root)
  }

  const updateDependency = async (node: ListNode | undefined): Promise<void> => {
    const tagged = readFolderData(node?.data)
    const args = tagged === undefined ? undefined : updateArguments(tagged.data)
    const state = tagged === undefined ? undefined : folders.get(tagged.root)
    if (tagged === undefined || args === undefined || state?.location === undefined) return
    if (!isWorkspaceTrusted()) {
      void vscode.window.showWarningMessage('MonoSmith: updating dependencies needs a trusted workspace.')

      return
    }
    const choice = await vscode.window.showWarningMessage(
      `MonoSmith: update ${tagged.data} to its latest version?`,
      { modal: true, detail: 'smith rewrites the manifests of every project that declares it, then installs again. It runs in a terminal so you can read what it does.' },
      'Update',
    )
    if (choice === 'Update') openSmithTerminal(state.location.executable, args, state.root)
  }

  const showInsight = async (node: TreeNode | undefined, kind: 'explain' | 'graph'): Promise<void> => {
    const picked = node?.kind === 'project' && node.project !== undefined
      ? { root: node.root, project: node.project }
      : await pickProject(readyWorkspaces(), kind === 'explain' ? 'Explain a project' : 'Show the graph of a project')
    const session = picked === undefined ? undefined : folders.get(picked.root)?.session
    if (picked === undefined || session === undefined) return
    const text = kind === 'explain' ? await readExplain(session.connection, picked.project) : await readGraph(session.connection, picked.project)
    await insight.show(`${picked.project}-${kind}.md`, text)
  }

  const setFilter = (next: ViewFilter): void => {
    tree.setFilter(next)
    void context.workspaceState.update(FILTER_KEY, next)
    void setContextKey('monosmith.affectedOnly', next.affectedOnly)
    void setContextKey('monosmith.filtered', next.text.trim() !== '')
  }

  /** Asks the doctor whether the folder has the agent files, shows the line in the view, and offers them once. */
  const checkAgentFiles = async (state: FolderState, offer: boolean): Promise<void> => {
    const session = state.session
    if (session === undefined) return
    try {
      agentFiles.set(state.root, judgeAgentFiles(await runDoctor(session.connection)))
    } catch (error) {
      note(`MonoSmith: could not read the agent files check of ${state.name}: ${error instanceof Error ? error.message : String(error)}`)

      return
    }
    sync()
    const asked = `${AGENT_ASKED_KEY}:${state.root}`
    if (!offer || !shouldOfferAgent({ state: agentFiles.get(state.root) ?? 'unknown', alreadyAsked: context.workspaceState.get<boolean>(asked) === true })) return
    await context.workspaceState.update(asked, true)
    const answer = await offerAgent(state.name)
    if (answer === 'install') await installIn(state, 'agent', false)
    else if (answer === 'never') await declineAgent(state)
  }

  const declineAgent = async (state: FolderState): Promise<void> => {
    if (state.location === undefined) return
    const location = state.location
    const failure = await optOutOfAgent(async (args) => await runSmith(location.executable, args, state.root))
    if (failure !== undefined) {
      void vscode.window.showWarningMessage(`MonoSmith: smith.toml was not changed. ${failure}`)

      return
    }
    agentFiles.set(state.root, 'unknown')
    sync()
    void vscode.window.showInformationMessage(`MonoSmith: ${state.name} will not be offered the agent again. smith.toml has enabled = false under [agent]; remove it to be asked.`)
  }

  const installIn = async (state: FolderState, scope: InstallScope, confirmed: boolean): Promise<boolean> => {
    if (state.location === undefined) return false
    if (!isWorkspaceTrusted()) {
      void vscode.window.showWarningMessage('MonoSmith: adding the agent writes files and needs a trusted workspace.')

      return false
    }
    const location = state.location
    const outcome = await installAgent({
      run:     async (args) => await runSmith(location.executable, args, state.root),
      confirm: async (preview) => await confirmInstall(preview, scope, state.name),
    }, scope, confirmed)
    note(`MonoSmith: install ${scope} in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellInstallOutcome(outcome, scope, state.root)
    if (outcome.kind === 'written' || outcome.kind === 'current') await checkAgentFiles(state, false)

    return outcome.kind === 'written' || outcome.kind === 'current'
  }

  /** The folder a command works in: the one in the row it came from, else the one the person picks. */
  const folderFor = async (argument: unknown): Promise<FolderState | undefined> => {
    const root = (argument as { root?: unknown } | undefined)?.root
    const given = typeof root === 'string' ? folders.get(root) : undefined
    if (given?.session !== undefined) return given
    const state = await folderOf()
    if (state === undefined) void vscode.window.showInformationMessage('MonoSmith: smith is not running in a folder, so there is nothing to do yet. Open a folder with a smith.toml and wait for the MonoSmith view to say it is ready.')

    return state
  }

  const installFiles = async (scope: InstallScope, argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state !== undefined) await installIn(state, scope, (argument as { confirmed?: unknown } | undefined)?.confirmed === true)
  }

  const talkToTheAgent = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined) return
    const location = state.location
    await talkToAgent({
      hasClaude:       hasClaudeCode,
      hasAgentFile:    () => existsSync(join(state.root, AGENT_FILE)),
      offerInstall:    async () => await installIn(state, 'agent', false) && existsSync(join(state.root, AGENT_FILE)),
      openAgent:       () => { openSmithTerminal(location.executable, ['agent'], state.root) },
      explainNoClaude: explainNoClaudeCode,
    })
  }

  const copyInstallCommand = async (): Promise<void> => {
    await vscode.env.clipboard.writeText(INSTALL_COMMAND)
    void vscode.window.showInformationMessage(`MonoSmith: copied "${INSTALL_COMMAND}". Run it in the workspace, then retry.`)
  }

  /** Asks which of several folders, or takes the only one. */
  const chooseFolder = async <T extends { root: string, name: string }>(candidates: readonly T[]): Promise<T | undefined> => {
    if (candidates.length <= 1) return candidates[0]
    const picked = await vscode.window.showQuickPick(candidates.map((candidate) => ({ label: candidate.name, description: candidate.root, candidate })), { placeHolder: 'Which folder?' })

    return picked?.candidate
  }

  /** The smith a terminal in the folder would run, or undefined after saying it is not installed. */
  const smithForTerminal = async (root: string): Promise<string | undefined> => {
    const found = locateSmith({
      setting:       vscode.workspace.getConfiguration('monosmith').get<string>('path', ''),
      workspaceRoot: root,
      trusted:       isWorkspaceTrusted(),
      platform:      process.platform,
      pathVariable:  process.env.PATH ?? '',
      exists:        existsSync,
    })
    if (found.location === undefined) {
      const copy = `Copy "${INSTALL_GLOBAL_COMMAND}"`
      const choice = await vscode.window.showWarningMessage('MonoSmith: smith is not installed on this computer. Install it, then try again.', copy)
      if (choice === copy) await vscode.env.clipboard.writeText(INSTALL_GLOBAL_COMMAND)

      return undefined
    }

    return found.location.executable
  }

  /**
   * Opens smith in a terminal on the folder to bring in: the repositories it
   * recognises, else the folders with files. Nothing is written by this
   * click: smith asks first.
   */
  const adopt = async (action: AdoptAction): Promise<void> => {
    const candidates = offers.length > 0
      ? offers.map((offer) => ({ root: offer.root, name: offer.name, why: `found ${offer.marker}`, project: offer.project }))
      : blanks.filter((folder) => folder.kind === 'unrecognised').map((folder) => ({ root: folder.root, name: folder.name, why: 'no tool it recognises', project: undefined }))
    const chosen = await chooseFolder(candidates)
    if (chosen === undefined) return
    // One project at the root: ask where it goes (smith never picks), then run the step that does not ask again.
    const answer = action === 'start' && chosen.project !== undefined ? await askWhereToMove(chosen.project) : undefined
    if (answer === undefined && action === 'start' && chosen.project !== undefined) return
    const executable = await smithForTerminal(chosen.root)
    if (executable === undefined) return
    note(`MonoSmith: ${action === 'start' ? 'bringing' : 'previewing bringing'} ${chosen.name} into MonoSmith (${chosen.why}).`)
    const terminal = openSmithTerminal(executable, answer === undefined ? toAdoptArguments(action) : toMoveArguments(answer.dir, answer.commit), chosen.root)
    if (action === 'start' && (answer === undefined || answer.commit)) offerUndoWhenDone(terminal, `${chosen.name} is converted`)
  }

  /**
   * Offers Undo when the command in the terminal ends well. It needs the
   * shell integration of the terminal; without it the offer is not made, and
   * "Undo the last change" stays in the palette.
   */
  function offerUndoWhenDone (terminal: vscode.Terminal, what: string): void {
    const ended = vscode.window.onDidEndTerminalShellExecution((event) => {
      if (event.terminal !== terminal) return
      ended.dispose()
      if (event.exitCode !== 0) return
      void vscode.window.showInformationMessage(`MonoSmith: ${what}. If it is not what you wanted, undo it in one step.`, 'Undo the last change').then(async (choice) => {
        if (choice !== undefined) await vscode.commands.executeCommand('monosmith.undo')
      })
    })
    context.subscriptions.push(ended)
  }

  /** Says a command that writes files needs a trusted workspace; true when it may go on. */
  const mayWrite = (what: string): boolean => {
    if (isWorkspaceTrusted()) return true
    void vscode.window.showWarningMessage(`MonoSmith: ${what} writes files and needs a trusted workspace. Trust the workspace, then try again.`)

    return false
  }

  /** Asks what to add, shows the plan of `smith add --dry-run`, and adds on a yes. */
  const addProjectHere = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('adding a project')) return
    output.show(true)
    const outcome = await addProject({
      run:   smithIn(state),
      apply: async (args) => await vscode.window.withProgress(
        { location: vscode.ProgressLocation.Notification, title: `MonoSmith: adding to ${state.name}… smith installs and tests the new project.` },
        async () => await smithIn(state, true, 600_000)(args),
      ),
      ask:     async (options) => await askAddRequest(options, state.snapshot?.projects.map((project) => project.name) ?? [], state.snapshot?.projects ?? []),
      confirm: confirmAddPlan,
    })
    note(`MonoSmith: add in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellAddOutcome(outcome, note)
    if (outcome.kind === 'added') await refreshFolder(state.root, 'full', false)
  }

  /** The folder and file a derive or why command works on: a Health row, a file given by the explorer or the editor, or the open editor. */
  const fileTarget = (argument: unknown): { state: FolderState, file: string } | undefined => {
    const tagged = readFolderData((argument as ListNode | undefined)?.data)
    if (tagged !== undefined) {
      const state = folders.get(tagged.root)

      return state === undefined ? undefined : { state, file: tagged.data }
    }
    const uri = argument instanceof vscode.Uri ? argument : vscode.window.activeTextEditor?.document.uri
    if (uri?.scheme !== 'file') return undefined
    const root = folderOfFile(ready().map((state) => state.root), uri.fsPath)
    const state = root === undefined ? undefined : folders.get(root)
    const file = root === undefined ? undefined : toFolderPath(root, uri.fsPath)

    return state === undefined || file === undefined ? undefined : { state, file }
  }

  const deriveFile = async (argument: unknown, action: 'explain' | 'apply' | 'adopt' | 'eject'): Promise<void> => {
    const target = fileTarget(argument)
    if (target === undefined) {
      void vscode.window.showInformationMessage('MonoSmith: choose a derived file first. Right-click it in the Health view, or open it in the editor and run the command again.')

      return
    }
    const { state, file } = target
    const run = smithIn(state)
    const explain = async (): Promise<void> => { await insight.show(`${basename(file)}-derive.md`, toExplainDocument(file, await explainDerivedFile(run, file))) }
    if (action === 'explain') {
      await explain()

      return
    }
    if (!mayWrite('changing a derived file')) return
    const status = deriveReports.get(state.root)?.files.find((item) => item.path === file)?.status
    if (action === 'apply' && status === undefined) {
      void vscode.window.showInformationMessage(`MonoSmith: smith has nothing to write for ${file}. Refresh the Health view to read the derived files again.`)

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
      if (answer === 'preview') await insight.show('derive-preview.md', toTextDocument('Derived files: what smith derive would write', await previewDerive(run)))
      else await explain()
    }
    const result = await changeDerivedFile(run, change, change === 'apply' ? undefined : file, commit)
    const printed = result.ok ? result.output.split(/\r?\n/).filter(Boolean) : []
    for (const line of printed) note(line)
    tellDeriveChange(change, change === 'apply' ? undefined : file, result.ok, result.ok ? '' : result.message)
    await health.reload()
    await refreshFolder(state.root, 'full', true)
  }

  /** Shows why a file is as it is, why a task runs again, or how a project resolved. */
  const showWhy = async (argument: unknown): Promise<void> => {
    const node = typeof argument === 'object' && argument !== null && 'kind' in argument ? argument as TreeNode : undefined
    const whyTask = async (root: string, project: string, task: string): Promise<void> => {
      const state = folders.get(root)
      if (state !== undefined) await insight.show(`${project}-${task}-why.md`, await readTaskWhy(smithIn(state), project, task))
    }
    if (node?.kind === 'task' && node.project !== undefined && node.task !== undefined) return await whyTask(node.root, node.project, node.task)
    if (node?.kind === 'project') return await showInsight(node, 'explain')

    const file = fileTarget(argument)
    // A file given by the explorer or the editor menu is answered at once; from the palette the person says what to ask about.
    const kinds = [
      ...(file === undefined ? [] : [{ label: '$(file) This file', description: file.file, pick: 'file' as const }]),
      { label: '$(package) A project', description: 'how its tasks resolved and where each setting came from', pick: 'project' as const },
      { label: '$(play) A task', description: 'why it runs again instead of coming from the cache', pick: 'task' as const },
      { label: '$(settings-gear) A setting', description: 'why a setting of smith.toml has its value, and why its default is what it is', pick: 'setting' as const },
    ]
    const chosen = argument instanceof vscode.Uri ? undefined : await vscode.window.showQuickPick(kinds, { title: 'MonoSmith: Why', placeHolder: 'What do you want to know about?' })
    switch (argument instanceof vscode.Uri ? 'file' : chosen?.pick) {
      case 'file': {
        if (file !== undefined) await insight.show(`${basename(file.file)}-why.md`, await readFileWhy(smithIn(file.state), file.file))
        break
      }
      case 'project': {
        await showInsight(undefined, 'explain')
        break
      }
      case 'setting': {
        const state = await folderFor(argument)
        const setting = state === undefined ? undefined : await vscode.window.showInputBox({ title: 'MonoSmith: Why', prompt: 'Which setting of smith.toml?', placeHolder: 'workspace.commit_operations' })
        if (state !== undefined && setting !== undefined && setting.trim() !== '') await insight.show(`${setting.trim()}-why.md`, await readSettingWhy(smithIn(state), setting.trim()))
        break
      }
      case 'task': {
        const picked = await pickTask(readyWorkspaces())
        if (picked?.target.kind === 'task') await whyTask(picked.root, picked.target.project, picked.target.task)
        else if (picked !== undefined) void vscode.window.showInformationMessage('MonoSmith: choose one task of a project, not all of its tasks.')
        break
      }
      default: {
        break
      }
    }
  }

  /** Asks which choice of which project to swap, shows the plan of `smith swap --dry-run`, and swaps on a yes. */
  const swapHere = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('swapping a choice')) return
    const node = typeof argument === 'object' && argument !== null && 'kind' in argument ? argument as TreeNode : undefined
    const given = node?.kind === 'project' ? node.project : undefined
    output.show(true)
    const outcome = await swapChoice({
      run:   smithIn(state),
      apply: async (args) => await vscode.window.withProgress(
        { location: vscode.ProgressLocation.Notification, title: `MonoSmith: swapping in ${state.name}… smith installs and tests the project.` },
        async () => await smithIn(state, true, 600_000)(args),
      ),
      parse:   parseChoices,
      ask:     async (choices) => await askSwapRequest(choices, state.snapshot?.projects ?? [], given),
      confirm: confirmSwapPlan,
    })
    note(`MonoSmith: swap in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellSwapOutcome(outcome, note)
    if (outcome.kind === 'swapped') await refreshFolder(state.root, 'full', false)
  }

  /** Lists where the workspace chose differently from smith's defaults (`smith divergences`), with the reason and the way back. */
  const showDivergences = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined) return
    const result = await smithIn(state)(['divergences'])
    if (result.code !== 0) {
      void vscode.window.showWarningMessage(`MonoSmith: smith could not list the divergences. ${describeFailure(result)}`)

      return
    }
    await insight.show('divergences.md', toTextDocument('Divergences from smith\'s defaults', result.stdout))
  }

  /** Shows what the last change smith committed did, and reverts it on a yes. */
  const undoHere = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('undoing a change')) return
    const outcome = await undoLastChange({ run: smithIn(state), confirm: async (plan) => await confirmUndo(plan, state.name) })
    note(`MonoSmith: undo in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellUndoOutcome(outcome)
    if (outcome.kind === 'undone') await refreshFolder(state.root, 'full', false)
  }

  /** Writes the HTML report of the workspace to a temp file and opens it in the browser. */
  const openReport = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined) return
    const outcome = await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: `MonoSmith: writing the report of ${state.name}…` },
      async () => await writeReport(smithIn(state), reportFile(tmpdir(), state.name)),
    )
    if (!outcome.ok) {
      void vscode.window.showWarningMessage(`MonoSmith: the report was not written. ${outcome.message}`)

      return
    }
    note(`MonoSmith: report of ${state.name}: ${outcome.file}`)
    await vscode.env.openExternal(vscode.Uri.file(outcome.file))
  }

  /** Writes AGENTS.md, from the agent menu or from the notice after the agent install. */
  const writeAgentsFile = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('AGENTS.md')) return
    const outcome = await writeAgentFile({ run: smithIn(state), confirm: async (lines) => await confirmAgentFile(lines, state.name) })
    note(`MonoSmith: AGENTS.md in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellAgentFileOutcome(outcome, state.root)
  }

  /** Runs a smith command in a temporary copy (`smith try`), shows what it did, and runs it for real on a yes. */
  const tryHere = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('running a command for real')) return
    const outcome = await tryCommand({
      run:   smithIn(state),
      ask:   askTryCommand,
      show:  async (page) => { await insight.show('try.md', page) },
      offer: offerTryApply,
    })
    note(`MonoSmith: try in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellTryOutcome(outcome)
    if (outcome.kind === 'tried' && outcome.applied) await refreshFolder(state.root, 'full', false)
  }

  /** Writes the launcher (`smith bootstrap`) that makes a laptop and CI run the same smith. */
  const bootstrapHere = async (argument: unknown): Promise<void> => {
    const state = await folderFor(argument)
    if (state?.location === undefined || !mayWrite('the smith launcher')) return
    const outcome = await writeLauncher({ run: smithIn(state), confirm: async () => await confirmLauncher(state.name) })
    note(`MonoSmith: launcher in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellLauncherOutcome(outcome)
    if (outcome.kind === 'written') await refreshFolder(state.root, 'full', false)
  }

  /** Opens `smith new --here` in a terminal on an empty folder. Nothing is written by this click: smith asks first and refuses a folder with files. */
  const createWorkspace = async (): Promise<void> => {
    const chosen = await chooseFolder(blanks.filter((folder) => folder.kind === 'empty'))
    if (chosen === undefined) return
    const executable = await smithForTerminal(chosen.root)
    if (executable === undefined) return
    note(`MonoSmith: creating a new workspace in ${chosen.name} (the folder is empty).`)
    openSmithTerminal(executable, toNewArguments(), chosen.root)
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
    command('monosmith.doctor', doctor),
    command('monosmith.showMenu', async () => {
      await showStatusMenu({ running: activeRun !== undefined, hasLastRun: lastRun.last() !== undefined, failed: failedTasks(), view: judgeViewState(states().map((state) => state.phase), withoutSmith()) })
    }),
    command('monosmith.showOutput', async () => { output.show() }),
    command('monosmith.openSettings', async () => { await vscode.commands.executeCommand('workbench.action.openSettings', 'monosmith') }),
    command('monosmith.copyInstallCommand', copyInstallCommand),
    command('monosmith.new', createWorkspace),
    command('monosmith.adopt.start', async () => { await adopt('start') }),
    command('monosmith.adopt.preview', async () => { await adopt('preview') }),
    command<unknown>('monosmith.installAgent', async (argument) => { await installFiles('agent', argument) }),
    command<unknown>('monosmith.installMcp', async (argument) => { await installFiles('mcp', argument) }),
    command<unknown>('monosmith.talkToAgent', talkToTheAgent),
    command<unknown>('monosmith.agentFile', writeAgentsFile),
    command<unknown>('monosmith.addProject', addProjectHere),
    command<unknown>('monosmith.undo', undoHere),
    command<unknown>('monosmith.swap', swapHere),
    command<unknown>('monosmith.divergences', showDivergences),
    command<unknown>('monosmith.try', tryHere),
    command<unknown>('monosmith.bootstrap', bootstrapHere),
    command<unknown>('monosmith.openReport', openReport),
    command<unknown>('monosmith.why', showWhy),
    command<unknown>('monosmith.derive.explain', async (argument) => { await deriveFile(argument, 'explain') }),
    command<unknown>('monosmith.derive.apply', async (argument) => { await deriveFile(argument, 'apply') }),
    command<unknown>('monosmith.derive.adopt', async (argument) => { await deriveFile(argument, 'adopt') }),
    command<unknown>('monosmith.derive.eject', async (argument) => { await deriveFile(argument, 'eject') }),
    command('monosmith.health.refresh', async () => { await health.reload() }),
    command('monosmith.health.fix', fixHealth),
    command('monosmith.dependencies.refresh', async () => { await dependencies.reload() }),
    command('monosmith.dependencies.up', runUp),
    command<ListNode>('monosmith.dependencies.update', updateDependency),
    command('monosmith.release.refresh', async () => { await release.reload() }),
    command('monosmith.explainProject', async (node) => { await showInsight(node, 'explain') }),
    command('monosmith.showGraph', async (node) => { await showInsight(node, 'graph') }),
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
