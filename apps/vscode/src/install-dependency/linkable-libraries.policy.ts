import type { LinkNode } from './install-request.contract.ts'

/**
 * Says whether a project may depend on a library, the way `smith install` decides: not itself, not an app,
 * the same language, and no cycle (the library must not already depend on the project, directly or not).
 * @param nodes - every project of the workspace
 * @param from - the project that would gain the dependency
 * @param library - the candidate
 * @returns true when smith would accept the link
 */
export function mayLink (nodes: readonly LinkNode[], from: string, library: string): boolean {
  const byName = new Map(nodes.map((node) => [node.name, node]))
  const a = byName.get(from)
  const b = byName.get(library)
  if (a === undefined || b === undefined || a.name === b.name || b.kind === 'app' || a.language !== b.language) return false
  const seen = new Set<string>([b.name])
  const queue = [b.name]
  for (let next = queue.shift(); next !== undefined; next = queue.shift()) {
    if (next === a.name) return false
    const deps = byName.get(next)?.deps ?? []
    for (const dep of deps) {
      if (seen.has(dep)) {
        continue
      }

      seen.add(dep)
      queue.push(dep)
    }
  }

  return true
}

/**
 * The libraries a project may depend on, in the order given.
 * @param nodes - every project of the workspace
 * @param from - the project that would gain the dependency
 * @returns the valid libraries, none when nothing can be linked
 */
export function linkableLibraries (nodes: readonly LinkNode[], from: string): LinkNode[] {
  return nodes.filter((node) => mayLink(nodes, from, node.name))
}
