import { join } from 'node:path'
import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** How writing the report ended: the file it is in, or smith's words and fix. */
export type ReportOutcome = { ok: true, file: string } | { ok: false, message: string }

/**
 * The file the report of a folder is written to: in the temp folder, so the
 * workspace stays clean, with a name that is safe on every system.
 * @param tempDir - the temp folder of the system
 * @param folderName - the name of the workspace folder
 */
export function reportFile (tempDir: string, folderName: string): string {
  return join(tempDir, `monosmith-report-${folderName.replaceAll(/[^\w.-]+/g, '-') || 'workspace'}.html`)
}

/**
 * Asks smith for the report page of the workspace (`smith report -o <file>`):
 * the map of the projects, health, derived-file drift and choices.
 * @param run - runs smith in the workspace folder
 * @param file - where to write the page
 * @returns the file, or what smith said
 */
export async function writeReport (run: (args: string[]) => Promise<SmithRun>, file: string): Promise<ReportOutcome> {
  const result = await run(['report', '-o', file])

  return result.code === 0 ? { ok: true, file } : { ok: false, message: describeFailure(result) }
}
