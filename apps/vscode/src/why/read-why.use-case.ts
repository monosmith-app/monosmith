import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { toFileWhyDocument, toTaskWhyDocument } from './why-document.mapper.ts'

/** Runs smith in the workspace folder with these arguments. */
export type RunSmithIn = (args: string[]) => Promise<SmithRun>

function pageOf (run: SmithRun, toPage: (json: string) => string): string {
  if (run.code !== 0) throw new Error(describeFailure(run))
  try {
    return toPage(run.stdout)
  } catch {
    throw new Error('smith answered in a form this editor does not read. Update the extension, or run the command in a terminal.')
  }
}

/**
 * Asks smith who owns a file and how to change it (`smith why <file> --json`).
 * @param run - runs smith in the folder
 * @param file - the file, relative to the folder
 * @returns a Markdown page
 * @throws Error with smith's words and the fix, when smith has no answer
 */
export async function readFileWhy (run: RunSmithIn, file: string): Promise<string> {
  return pageOf(await run(['why', '--json', '--', file]), toFileWhyDocument)
}

/**
 * Asks smith why a task runs again instead of coming from the cache (`smith why-miss <project>:<task> --json`).
 * @param run - runs smith in the folder
 * @param project - the project name
 * @param task - the task name
 * @returns a Markdown page
 * @throws Error with smith's words and the fix, when smith has no answer
 */
export async function readTaskWhy (run: RunSmithIn, project: string, task: string): Promise<string> {
  return pageOf(await run(['why-miss', `${project}:${task}`, '--json']), toTaskWhyDocument)
}
