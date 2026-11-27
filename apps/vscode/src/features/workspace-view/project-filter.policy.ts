/** What the filter of the Projects view reads of a project. */
export interface FilterProject {
  name:     string
  kind:     string
  language: string
  tasks:    readonly string[]
  /** The changes affect it. */
  affected: boolean
}

/** What a filter keeps of a project: the tasks to show, and whether the project must open to show them. */
export interface FilteredProject {
  tasks:  readonly string[]
  /** True when only some tasks matched, so the row opens on them. */
  reveal: boolean
}

const CHANGED = 'changed'

/**
 * Whether one word of the filter matches a project, as the terminal interface's filter reads it:
 * `kind:<prefix>`, `lang:<prefix>`, `changed` for a project the changes affect, or any part of the
 * name, the kind or the language.
 * @param project - the project
 * @param word - one word of the filter, in lower case
 * @returns true when the word matches the project itself
 */
export function matchesWord (project: FilterProject, word: string): boolean {
  if (word.startsWith('kind:')) return project.kind.toLowerCase().startsWith(word.slice('kind:'.length))
  if (word.startsWith('lang:')) return project.language.toLowerCase().startsWith(word.slice('lang:'.length))
  if (word === CHANGED && project.affected) return true

  return [project.name, project.kind, project.language].some((text) => text.toLowerCase().includes(word))
}

/**
 * Applies the filter of the Projects view to a project. Every word must match: the project
 * itself, or else one of its tasks, which then are the only tasks shown.
 * @param project - the project
 * @param text - the filter as typed; empty keeps everything
 * @returns what to show, or undefined when the project does not match
 */
export function filterProject (project: FilterProject, text: string): FilteredProject | undefined {
  const words = text.toLowerCase().split(/\s+/).filter(Boolean)
  const loose = words.filter((word) => !matchesWord(project, word))
  if (loose.length === 0) return { tasks: project.tasks, reveal: false }
  if (loose.some((word) => word.includes(':'))) return undefined
  const tasks = project.tasks.filter((task) => loose.every((word) => task.toLowerCase().includes(word)))

  return tasks.length === 0 ? undefined : { tasks, reveal: true }
}
