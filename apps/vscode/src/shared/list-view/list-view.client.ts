import * as vscode from 'vscode'
import type { ListNode, ListResult } from './list-node.contract.ts'
import { say } from '../words/index.ts'

/** A view of rows that are read when the view is looked at, and again on request. */
export interface ListView extends vscode.Disposable {
  /** Reads the rows now, behind the view's progress bar. */
  reload:     () => Promise<void>
  /** Reads the rows again without the progress bar, for rows that come from what the extension already holds. */
  redraw:     () => Promise<void>
  /** Says the rows are out of date: they are read again at once if the view is open, else when it opens. */
  invalidate: () => void
  /** The row that is selected. */
  selected:   () => ListNode | undefined
  /** Every row that is selected, for a view that lets several be selected. */
  selection:  () => readonly ListNode[]
  /**
   * Shows the view and the row with this id, without taking the keyboard from the editor.
   * @param id - the id of the row; the view alone shows when no row has it
   */
  reveal:     (id: string) => Promise<void>
}

/** What a list view needs. */
export interface ListViewOptions {
  /** The id the view has in `package.json`. */
  id:             string
  /** The name it shows in its header. */
  title:          string
  /** Reads the rows. A rejection becomes the message of the view. */
  load:           () => Promise<ListResult>
  /** Called after each load, with what the view now shows. */
  loaded?:        (result: ListResult) => void
  /** Lets the person select several rows (Ctrl or Shift and a click); a command then gets them all. */
  canSelectMany?: boolean
}

/**
 * Creates a tree view over rows that `load` reads. The rows are read lazily,
 * the first time the view is visible, so a registry query or a doctor run does
 * not slow down starting the editor.
 * @param options - the view's id, title and how to read its rows
 * @returns the view
 */
export function createListView (options: ListViewOptions): ListView {
  const changed = new vscode.EventEmitter<ListNode | undefined>()
  let nodes: ListNode[] = []
  let parents = new Map<string, ListNode>()
  let byId = new Map<string, ListNode>()
  let stale = true
  let generation = 0

  const index = (rows: readonly ListNode[], parent: ListNode | undefined): void => {
    for (const row of rows) {
      byId.set(row.id, row)
      if (parent !== undefined) parents.set(row.id, parent)
      index(row.children ?? [], row)
    }
  }

  const provider: vscode.TreeDataProvider<ListNode> = {
    onDidChangeTreeData: changed.event,
    getChildren:         (node) => (node === undefined ? nodes : (node.children ?? [])),
    getParent:           (node) => parents.get(node.id),
    getTreeItem:         (node) => {
      const hasChildren = (node.children?.length ?? 0) > 0
      const state = hasChildren
        ? (node.expanded === true ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.Collapsed)
        : vscode.TreeItemCollapsibleState.None
      const item = new vscode.TreeItem(node.label, state)
      item.id = node.id
      if (node.description !== undefined) item.description = node.description
      if (node.tooltip !== undefined) item.tooltip = node.tooltip
      if (node.context !== undefined) item.contextValue = node.context
      if (node.icon !== undefined) item.iconPath = new vscode.ThemeIcon(node.icon, node.color === undefined ? undefined : new vscode.ThemeColor(node.color))
      if (node.command !== undefined) item.command = { command: node.command, title: node.label, arguments: [node] }
      item.accessibilityInformation = { label: [node.label, node.description].filter(Boolean).join(', ') }

      return item
    },
  }
  const view = vscode.window.createTreeView<ListNode>(options.id, { treeDataProvider: provider, canSelectMany: options.canSelectMany === true })
  view.title = options.title

  const read = async (mine: number): Promise<void> => {
    let result: ListResult
    try {
      result = await options.load()
    } catch (error) {
      result = { nodes: [], message: say('vscode.list-view.could-not-load', { part: error instanceof Error ? error.message : String(error) }) }
    }
    if (mine !== generation) return
    nodes = result.nodes
    parents = new Map()
    byId = new Map()
    index(nodes, undefined)
    view.message = result.message
    view.description = result.summary
    options.loaded?.(result)
    changed.fire(undefined)
  }
  const reload = async (): Promise<void> => {
    const mine = ++generation
    stale = false
    await vscode.window.withProgress({ location: { viewId: options.id } }, async () => { await read(mine) })
  }
  const redraw = async (): Promise<void> => {
    stale = false
    await read(++generation)
  }
  const invalidate = (): void => {
    stale = true
    if (view.visible) void reload()
  }
  const subscription = view.onDidChangeVisibility((event) => {
    if (stale && event.visible) void reload()
  })
  if (view.visible) void reload()

  return {
    reload,
    redraw,
    invalidate,
    selected:  () => view.selection[0],
    selection: () => view.selection,
    reveal:    async (id) => {
      const row = byId.get(id)
      if (row === undefined) {
        await vscode.commands.executeCommand(`${options.id}.focus`)

        return
      }
      await view.reveal(row, { select: false, focus: false, expand: true })
    },
    dispose: () => {
      subscription.dispose()
      view.dispose()
      changed.dispose()
    },
  }
}
