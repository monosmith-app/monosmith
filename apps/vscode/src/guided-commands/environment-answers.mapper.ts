/** The parts of a row of the Projects view that a button on an Environments row passes to a command. */
interface EnvironmentRow {
  kind?:    string
  project?: string
  label?:   string
}

/**
 * What a button on an Environments row already answers: the web app, and for the minus button of one
 * environment, that environment. The questions about them are not asked again.
 * @param id - the id of the guided command
 * @param argument - what the editor passed to the command
 * @returns the answers by question, empty when the command did not start on an Environments row
 */
export function environmentAnswers (id: string, argument: unknown): Record<string, string | string[]> {
  const node = argument as EnvironmentRow | undefined
  if (node?.project === undefined || (node.kind !== 'environments' && node.kind !== 'environment')) return {}
  if (id === 'environmentsAdd') return { project: node.project }

  return id === 'environmentsRemove' ? { project: node.project, ...(node.kind === 'environment' && node.label !== undefined && { names: [node.label] }) } : {}
}
