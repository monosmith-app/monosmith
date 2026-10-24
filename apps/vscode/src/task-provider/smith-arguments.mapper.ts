import type { SmithTaskDefinition } from './smith-task-definition.contract.ts'

const WORD = /^[^\s-]\S*$/

/**
 * The arguments of the `smith` command a task definition stands for. A task
 * or project is one argument and never an option, so a value that looks like a
 * flag is refused.
 * @returns the arguments, or undefined when the definition names no task or a value is not a word
 */
export function toSmithArguments (definition: Pick<SmithTaskDefinition, 'task' | 'project' | 'affected' | 'bail' | 'noCache' | 'parallel'>): string[] | undefined {
  if (!WORD.test(definition.task)) return undefined
  if (definition.project !== undefined && definition.project !== '' && !WORD.test(definition.project)) return undefined

  return [
    'run',
    definition.task,
    ...(definition.project === undefined || definition.project === '' ? [] : ['-p', definition.project]),
    ...(definition.affected === true ? ['--affected'] : []),
    ...(definition.bail === true ? ['--bail'] : []),
    ...(definition.noCache === true ? ['--no-cache'] : []),
    ...((definition.parallel ?? 0) > 0 ? ['--parallel', String(Math.floor(definition.parallel ?? 0))] : []),
  ]
}

/** What Run Task shows for a definition: `build`, `test (affected)`, `web:lint`. */
export function nameTask (definition: Pick<SmithTaskDefinition, 'task' | 'project' | 'affected'>): string {
  const where = definition.project === undefined || definition.project === '' ? definition.task : `${definition.project}:${definition.task}`

  return definition.affected === true ? `${where} (affected)` : where
}
