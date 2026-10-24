import type { GraphAnswer, GraphProject } from './insight-answer.contract.ts'

/**
 * Writes the neighbourhood of a project as a Markdown page with two trees in
 * plain text: what the project uses, to the bottom, and what uses it.
 * @param project - the project that was asked for
 * @param answer - `graph` focused on it
 */
export function toGraphDocument (project: string, answer: GraphAnswer): string {
  const listed = answer.projects ?? []
  const projects = new Map(listed.map((entry) => [entry.name, entry]))
  const dependents = new Map<string, string[]>()
  for (const entry of listed) {
    const deps = entry.deps ?? []
    for (const dep of deps) dependents.set(dep, [...(dependents.get(dep) ?? []), entry.name])
  }
  const uses = tree(project, (name) => projects.get(name)?.deps ?? [], projects)
  const usedBy = tree(project, (name) => dependents.get(name) ?? [], projects)
  const self = projects.get(project)

  return [
    `# Graph of ${project}`,
    '',
    `${describe(project, self)}: what it uses and what uses it. The same as \`smith graph --focus ${project}\`.`,
    '',
    '## Uses',
    '',
    uses.length === 0 ? 'Nothing: it depends on no other project.' : fence([project, ...uses]),
    '',
    '## Used by',
    '',
    usedBy.length === 0 ? 'Nothing: no other project depends on it.' : fence([project, ...usedBy]),
    '',
  ].join('\n')
}

function fence (lines: string[]): string {
  return ['```text', ...lines, '```'].join('\n')
}

function describe (name: string, entry: GraphProject | undefined): string {
  const facts = [entry?.kind, entry?.language].filter(Boolean).join(', ')

  return facts === '' ? `\`${name}\`` : `\`${name}\` (${facts})`
}

function label (name: string, projects: ReadonlyMap<string, GraphProject>): string {
  const entry = projects.get(name)
  const facts = [entry?.kind, entry?.language].filter(Boolean).join(', ')

  return facts === '' ? name : `${name} (${facts})`
}

/** The box-drawn lines of the tree under `root`; a project already on the path is marked, not followed. */
function tree (root: string, next: (name: string) => string[], projects: ReadonlyMap<string, GraphProject>): string[] {
  const walk = (name: string, prefix: string, path: readonly string[]): string[] => {
    const children = next(name)

    const lines: string[] = []
    for (const [index, child] of children.entries()) {
      const last = index === children.length - 1
      const branch = `${prefix}${last ? '└─ ' : '├─ '}`
      if (path.includes(child)) lines.push(`${branch}${child} (cycle)`)
      else lines.push(`${branch}${label(child, projects)}`, ...walk(child, `${prefix}${last ? ' '.repeat(3) : '│  '}`, [...path, child]))
    }

    return lines
  }

  return walk(root, '', [root])
}
