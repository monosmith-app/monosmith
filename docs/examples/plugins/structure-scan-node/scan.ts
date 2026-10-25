import { readdirSync, statSync } from 'node:fs'
import { basename, join } from 'node:path'
import { setTimeout as sleep } from 'node:timers/promises'
import type { Finding, Project, SmithPlugin } from '@monosmith/plugin-sdk'

/** The size from which a file is worth a warning. */
const BIG_FILE = 1 << 20

/** What a scan throws when it is cancelled: the SDK answers a cancelled request that fails with the request-cancelled code, whatever the message. */
const STOPPED = 'the scan was stopped'

/**
 * The files below `dir` over a mebibyte, relative to dir with slashes. It
 * checks the signal before every file (and pauses `delayMs` before each, which
 * stands in for a slow disk), so a cancel ends it within one file's time.
 */
export async function bigFiles (dir: string, delayMs: number, signal: AbortSignal): Promise<string[]> {
  const found: string[] = []
  const walk = async (relative: string): Promise<void> => {
    const entries = readdirSync(join(dir, relative), { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))
    for (const entry of entries) {
      const path = relative === '' ? entry.name : `${relative}/${entry.name}`
      if (entry.isDirectory()) {
        await walk(path)
        continue
      }
      try {
        await sleep(delayMs, undefined, { signal })
      } catch {
        throw new Error(STOPPED)
      }
      if (statSync(join(dir, path)).size > BIG_FILE) found.push(path)
    }
  }
  await walk('')

  return found
}

/**
 * The plugin: a structure plugin whose doctor walks every file of a project.
 * A walk can take a while, so it shows the two things a protocol client expects
 * of such a method:
 *
 * - it is `cancellable`: the SDK passes each method an `AbortSignal` that aborts
 *   when smith cancels the request, the walk watches it, and the plugin advertises
 *   the cancel capability;
 * - it declares a `concurrency` of 4: smith may send up to four requests at once,
 *   and the SDK runs them together. Scanning shares no state, so this costs nothing.
 */
export function createScanPlugin (delayMs = 0): SmithPlugin {
  let workspaceRoot = ''

  return {
    cancellable: true,

    describe: (params) => {
      workspaceRoot = params.workspaceRoot

      return {
        plugin:      { name: 'widget-scan', version: '0.1.0' },
        kind:        'convention',
        roles:       ['structure'],
        languages:   ['widget'],
        concurrency: 4,
      }
    },

    // No file of a project is bigger than a mebibyte.
    doctor: async ({ projects }, signal) => {
      const findings: Finding[] = []
      for (const project of projects) {
        const dir = join(workspaceRoot, ...project.root.split('/'))
        findings.push(await finding(project, dir, delayMs, signal))
      }

      return { findings }
    },
  }
}

async function finding (project: Project, dir: string, delayMs: number, signal: AbortSignal): Promise<Finding> {
  const check = `${project.name} has no file over 1 MiB`
  let big: string[]
  try {
    big = await bigFiles(dir, delayMs, signal)
  } catch (error) {
    if (error instanceof Error && error.message === STOPPED) throw error
    throw new Error(`scanning ${basename(dir)}: ${error instanceof Error ? error.message : String(error)}`, { cause: error })
  }
  if (big.length === 0) return { check, ok: true }

  return {
    check,
    ok:      false,
    detail:  big.join(', '),
    remedy:  `keep large files out of the repository: ${big[0]}`,
    warning: true, // reported, but does not fail smith doctor
  }
}
