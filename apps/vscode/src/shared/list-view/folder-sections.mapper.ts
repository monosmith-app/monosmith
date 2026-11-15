import type { ListNode } from './list-node.contract.ts'

/** What one folder of the editor contributes to a list view. */
export interface FolderRows {
  root:  string
  name:  string
  nodes: ListNode[]
}

/**
 * With one folder its rows as they are; with several, a row per folder that
 * holds that folder's rows, so the same finding in two folders reads apart.
 */
export function groupByFolder (folders: readonly FolderRows[]): ListNode[] {
  const first = folders[0]
  if (first !== undefined && folders.length === 1) return first.nodes

  return folders.map((folder): ListNode => ({
    id:       `folder:${folder.root}`,
    label:    folder.name,
    tooltip:  folder.root,
    icon:     'root-folder',
    expanded: true,
    children: folder.nodes,
  }))
}

const SEPARATOR = '\0'

/**
 * Marks every row (and the rows under it) that carries data with the folder it
 * belongs to, so a command given the row knows which smith to ask.
 */
export function tagWithFolder (nodes: readonly ListNode[], root: string): ListNode[] {
  return nodes.map((node) => ({
    ...node,
    ...(node.data !== undefined && { data: `${root}${SEPARATOR}${node.data}` }),
    ...(node.children !== undefined && { children: tagWithFolder(node.children, root) }),
  }))
}

/** Splits what `tagWithFolder` joined; undefined for data that was not tagged. */
export function readFolderData (data: string | undefined): { root: string, data: string } | undefined {
  const at = data?.indexOf(SEPARATOR) ?? -1

  return data === undefined || at === -1 ? undefined : { root: data.slice(0, at), data: data.slice(at + 1) }
}
