import type { DoctorReport } from '../health/index.ts'
import type { UpdatesReport } from '../dependencies/index.ts'
import type { ReleasePreview } from '../release-preview/index.ts'
import { agoText, runState, type RunRecord } from '../runs-view/index.ts'
import { formatDuration } from '../workspace-view/index.ts'
import type { FolderWorkspace } from '../../shared/workspace/index.ts'
import type { HomeFacts, HomeHealth, HomeRelease, HomeRun } from './home-facts.contract.ts'

/** What the extension holds, from which Home is drawn. */
export interface HomeSources {
  /** The folders whose projects arrived. */
  workspaces:   readonly FolderWorkspace[]
  /** How many folders run smith; Home counts as loaded once each of them has its projects. */
  folders:      number
  /** What doctor answered in each folder; undefined while it has not answered everywhere. */
  reports?:     readonly DoctorReport[]
  /** Why doctor could not run. */
  healthError?: string
  /** The release plan of each folder; undefined while it was not read. */
  previews?:    readonly ReleasePreview[]
  /** The updates of each folder; undefined while the registries were not asked. */
  updates?:     readonly UpdatesReport[]
  /** The newest run of the session. */
  last?:        RunRecord
  /** The time now, in milliseconds. */
  now:          number
}

/**
 * Turns what the extension holds into the facts of Home, summed over the folders.
 * @param sources - the snapshots, reports and runs the extension already has
 * @returns the facts
 */
export function toHomeFacts (sources: HomeSources): HomeFacts {
  const projects = sources.workspaces.flatMap((folder) => folder.snapshot.projects)
  const affected = sources.workspaces.reduce((sum, folder) => sum + folder.snapshot.projects.filter((project) => folder.snapshot.affected.has(project.name)).length, 0)
  const outdated = sources.updates?.reduce((sum, report) => sum + (report.updates ?? []).length, 0)
  const release = sources.previews === undefined ? undefined : toRelease(sources.previews)

  return {
    projectsLoaded: sources.folders > 0 && sources.workspaces.length >= sources.folders,
    projects:       projects.length,
    languages:      projects.map((project) => project.language),
    affected,
    ...(sources.reports !== undefined && { health: toHealth(sources.reports) }),
    ...(sources.healthError !== undefined && { healthError: sources.healthError }),
    ...(sources.last !== undefined && { last: toRun(sources.last, sources.now) }),
    ...(release !== undefined && { release }),
    releaseKnown:   sources.previews !== undefined,
    ...(outdated !== undefined && { outdated }),
  }
}

function toHealth (reports: readonly DoctorReport[]): HomeHealth {
  const findings = reports.flatMap((report) => report.findings ?? [])
  const failed = findings.filter((finding) => !finding.ok && finding.warning !== true)
  const warnings = findings.filter((finding) => !finding.ok && finding.warning === true)

  return {
    failed:       failed.length,
    warnings:     warnings.length,
    passed:       findings.filter((finding) => finding.ok).length,
    fixable:      [...failed, ...warnings].filter((finding) => finding.fix !== undefined).length,
    problems:     failed.map((finding) => finding.check),
    warningNotes: warnings.map((finding) => finding.check),
  }
}

function toRelease (previews: readonly ReleasePreview[]): HomeRelease | undefined {
  const results = previews.flatMap((preview) => preview.results ?? [])
  const [first] = results
  if (first === undefined) return undefined
  const count = (level: string): number => results.filter((result) => result.level === level).length

  return {
    packages: results.length,
    name:     first.project,
    from:     first.current ?? '',
    to:       first.next,
    commits:  first.commits?.length ?? 0,
    major:    count('major'),
    minor:    count('minor'),
    patch:    count('patch'),
  }
}

function toRun (run: RunRecord, now: number): HomeRun {
  const state = runState(run)
  const count = (status: string): number => run.tasks.filter((task) => task.status === status).length

  return {
    command: run.command,
    ...(run.tasks.find((task) => task.status === 'failed')?.task !== undefined && { task: run.tasks.find((task) => task.status === 'failed')?.task }),
    running: state === 'running',
    ok:      state === 'passed',
    stopped: state === 'stopped',
    took:    formatDuration((run.endedAt ?? now) - run.startedAt),
    ago:     run.endedAt === undefined ? '' : agoText(run.endedAt, now),
    passed:  count('passed'),
    cached:  count('cached'),
    failed:  count('failed'),
  }
}
