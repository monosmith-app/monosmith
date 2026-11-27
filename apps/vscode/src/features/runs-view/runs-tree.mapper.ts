import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import type { TaskStatus } from '../task-run/index.ts'
import { formatDuration, type ProjectRun } from '../workspace-view/index.ts'
import type { RunRecord, RunTask } from './run-history.store.ts'
import { say, sayCount, type WordKey } from '../../shared/words/index.ts'

/** How a run stands: going, passed, failed, or stopped by the person. */
export type RunState = 'running' | 'passed' | 'failed' | 'stopped'

/** The codicon and the theme colour of each task status. */
const MARKS: Record<TaskStatus, { icon: string, color?: string }> = {
  queued:  { icon: 'circle-outline' },
  running: { icon: 'sync~spin', color: 'charts.blue' },
  passed:  { icon: 'pass', color: 'charts.green' },
  cached:  { icon: 'pass', color: 'charts.green' },
  failed:  { icon: 'error', color: 'charts.red' },
  skipped: { icon: 'debug-step-over' },
}

const STATE_MARKS: Record<RunState, { icon: string, color?: string, word: WordKey }> = {
  running: { icon: 'sync~spin', color: 'charts.blue', word: 'runpage.state.running' },
  passed:  { icon: 'pass', color: 'charts.green', word: 'runpage.state.passed' },
  failed:  { icon: 'error', color: 'charts.red', word: 'runpage.state.failed' },
  stopped: { icon: 'circle-slash', word: 'runpage.state.stopped' },
}

/** How the Runs view is narrowed. */
export interface RunsFilter {
  /** Only the failed tasks (the terminal interface's `e`). */
  errorsOnly: boolean
}

/**
 * How a run stands. A run that ended with a task cancelled was stopped.
 * @param run - the run
 * @returns its state
 */
export function runState (run: RunRecord): RunState {
  if (run.endedAt === undefined) return 'running'
  if (run.ok === true) return 'passed'

  return run.tasks.some((task) => task.reason === 'cancelled') ? 'stopped' : 'failed'
}

/**
 * Says how long ago a time was, in the words of the terminal interface.
 * @param then - the time, in milliseconds
 * @param now - the time now, in milliseconds
 * @returns `just now`, `3 min ago` or `2 h ago`
 */
export function agoText (then: number, now: number): string {
  const minutes = Math.floor((now - then) / 60_000)
  if (minutes < 1) return say('tui.format_ago.just_now')
  if (minutes < 60) return say('tui.format_ago.min_ago', { int: minutes })

  return say('tui.format_ago.h_ago', { int: Math.floor(minutes / 60) })
}

/** How long a run took, or has taken so far. */
function took (run: RunRecord, now: number): string {
  return formatDuration((run.endedAt ?? now) - run.startedAt)
}

/** The tasks of a run that are done (ended in any way). */
function doneCount (run: RunRecord): number {
  return run.tasks.filter((task) => task.status !== 'queued' && task.status !== 'running').length
}

/** `✓ 3 passed · 2 cached · 1 failed`: the counts of a run with their marks, the ones that are not zero. */
function counts (run: RunRecord): string {
  const tally: Array<[TaskStatus, WordKey, string]> = [
    ['failed', 'runpage.counts.failed', '×'],
    ['passed', 'runpage.counts.passed', '✓'],
    ['cached', 'runpage.counts.cached', '≈'],
    ['running', 'runpage.counts.running', '◐'],
    ['queued', 'runpage.counts.queued', '○'],
    ['skipped', 'runpage.counts.skipped', '»'],
  ]

  return tally
    .map(([status, word, mark]) => ({ word, mark, n: run.tasks.filter((task) => task.status === status).length }))
    .filter(({ n }) => n > 0)
    .map(({ word, mark, n }) => say(word, { mark, n }))
    .join(' · ')
}

/** What a task row says beside its name: how long it ran, `hit` for a cached one, or why it did not pass. */
function taskNote (task: RunTask): string {
  if (task.status === 'cached') return say('runpage.cache.hit')
  if (task.reason !== undefined && task.status !== 'passed') return task.reason

  return (task.durationMs ?? 0) > 0 ? formatDuration(task.durationMs ?? 0) : ''
}

/** The project part of a task id: `web` of `web:build`. */
function projectOf (task: string): string {
  const colon = task.indexOf(':')

  return colon === -1 ? task : task.slice(0, colon)
}

/** The task part of a task id: `build` of `web:build`. */
function nameOf (task: string): string {
  const colon = task.indexOf(':')

  return colon === -1 ? task : task.slice(colon + 1)
}

/**
 * The tasks of a run grouped by project, in the order they were first seen, under `All output`.
 * Errors only keeps the failed tasks; a project without one leaves.
 */
function runChildren (run: RunRecord, filter: RunsFilter): ListNode[] {
  const shown = filter.errorsOnly ? run.tasks.filter((task) => task.status === 'failed') : run.tasks
  const projects = [...new Set(shown.map((task) => projectOf(task.task)))]
  const output: ListNode = { id: `run:${run.id}:output`, label: say('runpage.tree.all_output'), icon: 'output', command: 'monosmith.showOutput', tooltip: run.command }
  if (run.tasks.length === 0) return [output, { id: `run:${run.id}:none`, label: say('runpage.tree.no_tasks'), icon: 'info' }]
  if (shown.length === 0) return [output, { id: `run:${run.id}:clean`, label: say('runpage.tree.no_errors'), icon: 'pass', color: 'charts.green' }]

  return [output, ...projects.map((project): ListNode => ({
    id:       `run:${run.id}:project:${project}`,
    label:    project,
    icon:     'folder',
    expanded: true,
    children: shown.filter((task) => projectOf(task.task) === project).map((task): ListNode => ({
      id:          `run:${run.id}:task:${task.task}`,
      label:       nameOf(task.task),
      description: taskNote(task),
      icon:        MARKS[task.status].icon,
      ...(MARKS[task.status].color !== undefined && { color: MARKS[task.status].color }),
      tooltip:     [task.task, task.status, task.reason].filter(Boolean).join(' · '),
      context:     'run.task',
      command:     'monosmith.runs.showLog',
      data:        task.task,
    })),
  }))]
}

/**
 * The row of a run: its state and its tasks (`Failed 4 tasks`), then how far it is and how long
 * it took (`3/4 75% · 12s`), the command line in the tooltip.
 * @param run - the run
 * @param now - the time now, in milliseconds
 * @param filter - how the view is narrowed
 * @param open - whether the row shows its tasks
 * @returns the row with its tasks under it
 */
export function toRunRow (run: RunRecord, now: number, filter: RunsFilter, open: boolean): ListNode {
  const state = runState(run)
  const mark = STATE_MARKS[state]
  const total = run.tasks.length
  const done = doneCount(run)
  const bar = total === 0 ? '' : say('runpage.bar.label', { done, total, percent: Math.round((done / total) * 100) })

  return {
    id:          `run:${run.id}`,
    label:       say(mark.word, { what: sayCount('runpage.state.tasks', total) }),
    description: [bar, took(run, now)].filter(Boolean).join(' · '),
    icon:        mark.icon,
    ...(mark.color !== undefined && { color: mark.color }),
    tooltip:     [`$ ${run.command}`, counts(run), agoText(run.endedAt ?? run.startedAt, now)].filter(Boolean).join('\n'),
    context:     state === 'running' ? 'run.running' : 'run.ended',
    expanded:    open,
    children:    runChildren(run, filter),
  }
}

/**
 * Lays the runs of the session out as the Runs view shows them: the newest run open with its
 * tasks grouped by project, then **Runs this session** with the older ones, closed.
 * @param runs - the runs, the newest first
 * @param now - the time now, in milliseconds
 * @param filter - how the view is narrowed
 * @returns the rows, with the command line of the newest run beside the title
 */
export function toRunRows (runs: readonly RunRecord[], now: number, filter: RunsFilter): ListResult {
  const [newest, ...older] = runs
  if (newest === undefined) return { nodes: [], message: `${say('runpage.empty.title')}. ${say('runpage.empty.intro')}` }

  return {
    nodes: [
      toRunRow(newest, now, filter, true),
      ...(older.length === 0
        ? []
        : [{
            id:          'runs:session',
            label:       say('runpage.tree.session'),
            description: String(older.length),
            icon:        'history',
            children:    older.map((run) => toRunRow(run, now, filter, false)),
          }]),
    ],
    summary: newest.command,
  }
}

/** The worst status of some tasks: failed, then running, queued, skipped, passed, cached. */
export function worstStatus (statuses: readonly TaskStatus[]): TaskStatus | undefined {
  const order: TaskStatus[] = ['failed', 'running', 'queued', 'skipped', 'passed', 'cached']

  return order.find((status) => statuses.includes(status))
}

/**
 * The runs of the session that ran a task of a project, the newest first.
 * @param runs - the runs of the session, the newest first
 * @param root - the folder of the project
 * @param project - the project
 * @param now - the time now, in milliseconds
 * @returns one line for each such run
 */
export function projectRuns (runs: readonly RunRecord[], root: string, project: string, now: number): ProjectRun[] {
  return runs.flatMap((run): ProjectRun[] => {
    if (run.root !== root) return []
    const mine = run.tasks.filter((task) => projectOf(task.task) === project)
    const status = worstStatus(mine.map((task) => task.status))

    return status === undefined ? [] : [{ id: run.id, tasks: mine.map((task) => nameOf(task.task)).join(' '), status, note: [took(run, now), agoText(run.endedAt ?? run.startedAt, now)].join(' · ') }]
  })
}

/**
 * Whether the tests of the session passed: what the checks before a release say.
 * @param runs - the runs of the session, the newest first
 * @returns the state of the newest run that ended with a `test` task, or `none`
 */
export function testsState (runs: readonly RunRecord[]): 'passed' | 'failed' | 'none' {
  for (const run of runs) {
    if (run.endedAt === undefined) continue
    const tests = run.tasks.filter((task) => nameOf(task.task) === 'test')
    if (tests.length > 0) return tests.some((task) => task.status === 'failed') ? 'failed' : 'passed'
  }

  return 'none'
}
