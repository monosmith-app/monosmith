import type { ExplainedTask } from './insight-answer.contract.ts'

/**
 * Writes how a project's tasks resolved as a Markdown page: one section per
 * task with its description and a table of its fields, each with where the
 * value came from (the language plugin, a stack, a convention, `smith.toml`).
 */
export function toExplainDocument (project: string, tasks: readonly ExplainedTask[]): string {
  const sections = tasks.map((task) => {
    const fields = task.fields ?? []
    const table = fields.length === 0
      ? ''
      : [
          '',
          '',
          '| Field | Value | Comes from |',
          '| --- | --- | --- |',
          ...fields.map((field) => `| ${cell(field.name)} | ${code(field.value)} | ${cell(field.source)} |`),
        ].join('\n')

    return [`## ${task.id}`, task.description === undefined || task.description === '' ? '' : `\n\n${markup(task.description)}`, table].join('')
  })

  return [
    `# Explain ${project}`,
    '',
    `How the tasks of \`${project}\` resolved, and where each field came from. The same as \`smith explain ${project}\`.`,
    '',
    ...(tasks.length === 0 ? ['This project has no tasks.'] : sections.flatMap((section) => [section, ''])),
  ].join('\n').trimEnd() + '\n'
}

function cell (text: string): string {
  return text.replaceAll('|', String.raw`\|`).replaceAll(/\r?\n/g, ' ')
}

function code (text: string): string {
  const safe = cell(text)

  return safe === '' || safe === '—' ? safe : `\`${safe.replaceAll('`', "'")}\``
}

/** Text of smith that holds `<project dir>` and the like must not be read as a tag. */
function markup (text: string): string {
  return text.replaceAll('<', '&lt;').replaceAll('>', '&gt;')
}
