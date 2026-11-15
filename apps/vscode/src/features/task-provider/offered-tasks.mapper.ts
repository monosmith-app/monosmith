import type { OfferedTask } from './smith-task-definition.contract.ts'

const GROUPS: Record<string, OfferedTask['group']> = { build: 'build', test: 'test', clean: 'clean' }

/**
 * The tasks Run Task lists for a folder: each task name any project has, run
 * in every project that has it, and the same limited to the projects the
 * changes affect. A task of one project is written in tasks.json instead,
 * where the definition names the project.
 * @param taskNames - the task names of the folder's projects, with repeats
 */
export function toOfferedTasks (taskNames: readonly string[]): OfferedTask[] {
  return [...new Set(taskNames)].sort((a, b) => a.localeCompare(b)).flatMap((task): OfferedTask[] => [
    { name: task, definition: { type: 'smith', task }, group: GROUPS[task] },
    { name: `${task} (affected)`, definition: { type: 'smith', task, affected: true } },
  ])
}
