/** The parts of a row of the Projects view that the Settings entry of a project passes to a command. */
interface ProjectRow {
  kind?:    string
  project?: string
}

/**
 * What the Settings entry of a project row already answers: which project. That question is not asked again.
 * @param id - the id of the guided command
 * @param argument - what the editor passed to the command
 * @returns the answers by question, empty when the command did not start on a project row
 */
export function projectRowAnswers (id: string, argument: unknown): Record<string, string> {
  const node = argument as ProjectRow | undefined

  return id === 'projectSettings' && node?.kind === 'project' && node.project !== undefined ? { project: node.project } : {}
}
