import type { FolderState } from '../folder-session/index.ts'
import type { SmithConnection } from '../smith-client/index.ts'
import { groupByFolder, tagWithFolder } from './folder-sections.mapper.ts'
import type { ListResult } from './list-node.contract.ts'

/**
 * Reads a list view's rows from every folder that runs smith, and lays them out by folder.
 * @param ready - the folders that run smith now
 * @param nothing - what the view says when no folder runs smith
 * @param read - reads the rows of one folder; an error becomes the message of that folder
 * @returns the rows, with the messages and summaries of the folders
 */
export async function loadAcross (ready: () => FolderState[], nothing: string, read: (state: FolderState, connection: SmithConnection) => Promise<ListResult>): Promise<ListResult> {
  const targets = ready()
  if (targets.length === 0) return { nodes: [], message: nothing }
  const results = await Promise.all(targets.map(async (state) => {
    try {
      return { state, result: await read(state, (state.session as NonNullable<FolderState['session']>).connection) }
    } catch (error) {
      return { state, result: { nodes: [], message: `${state.name}: ${error instanceof Error ? error.message : String(error)}` } satisfies ListResult }
    }
  }))
  const nodes = groupByFolder(results.map(({ state, result }) => ({ root: state.root, name: state.name, nodes: tagWithFolder(result.nodes, state.root) })))
  const messages = results.flatMap(({ result }) => result.message ?? [])
  const summaries = results.flatMap(({ result }) => result.summary ?? [])

  return {
    nodes,
    ...(messages.length > 0 && (nodes.length === 0 || results.length === 1) && { message: messages.join(' ') }),
    ...(summaries.length > 0 && { summary: results.length === 1 ? summaries[0] : summaries.join(' · ') }),
  }
}
