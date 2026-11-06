import type { CommandHost } from '../command-host/index.ts'
import { pickProject } from '../quick-pick/index.ts'
import type { InsightDocuments } from './insight-documents.client.ts'
import { readExplain, readGraph } from './read-insight.use-case.ts'

/** The parts of a row of the Projects view that tell which project a command started on. */
export interface InsightRow {
  kind?:    string
  root:     string
  project?: string
}

/**
 * Shows how a project resolved (`explain`) or its place in the graph, as a page in the editor.
 * @param host - the folders and the Output channel of the extension
 * @param insight - where the pages are shown
 * @param node - the row the command started on, or undefined to pick a project
 * @param kind - what to show
 */
export async function showInsight (host: CommandHost, insight: InsightDocuments, node: InsightRow | undefined, kind: 'explain' | 'graph'): Promise<void> {
  const picked = node?.kind === 'project' && node.project !== undefined
    ? { root: node.root, project: node.project }
    : await pickProject(host.readyWorkspaces(), kind === 'explain' ? 'Explain a project' : 'Show the graph of a project')
  const session = picked === undefined ? undefined : host.registry.folders.get(picked.root)?.session
  if (picked === undefined || session === undefined) return
  const text = kind === 'explain' ? await readExplain(session.connection, picked.project) : await readGraph(session.connection, picked.project)
  await insight.show(`${picked.project}-${kind}.md`, text)
}
