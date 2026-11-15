/**
 * The project a command was started on, when the person used the menu of a project row.
 * @param argument - what the editor passed to the command
 * @returns the project name, or undefined for any other argument
 */
export function projectGiven (argument: unknown): string | undefined {
  const node = typeof argument === 'object' && argument !== null && 'kind' in argument ? argument as { kind: unknown, project?: string } : undefined

  return node?.kind === 'project' ? node.project : undefined
}
