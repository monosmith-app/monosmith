import * as vscode from 'vscode'
import type { ListNode, ListResult } from './list-node.contract.ts'

/** A view of rows that are read when the view is looked at, and again on request. */
export interface ListView extends vscode.Disposable {
  /** Reads the rows now, behind the view's progress bar. */
  reload:     () => Promise<void>
  /** Says the rows are out of date: they are read again at once if the view is open, else when it opens. */
  invalidate: () => void
  /** The row that is selected. */
  selected:   () => ListNode | undefined
}

/** What a list view needs. */
export interface ListViewOptions {
  /** The id the view has in `package.json`. */
  id:      string
  /** The name it shows in its header. */
  title:   string
  /** Reads the rows. A rejection becomes the message of the view. */
  load:    () => Promise<ListResult>
  /** Called after each load, with what the view now shows. */
  loaded?: (result: ListResult) => void
}

/**
 * Creates a tree view over rows that `load` reads. The rows are read lazily,
 * the first time the view is visible, so a registry query or a doctor run does
 * not slow down starting the editor.
 */
export function createListView (options: ListViewOptions): ListView {
  const changed = new vscode.EventEmitter<ListNode | undefined>()
  let nodes: ListNode[] = []
  let stale = true
  let generation = 0

  const provider: vscode.TreeDataProvider<ListNode> = {
    onDidChangeTreeData: changed.event,
    getChildren:         (node) => (node === undefined ? nodes : (node.children ?? [])),
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
      item.accessibilityInformation = { label: [node.label, node.description].filter(Boolean).join(', ') }

      return item
    },
  }
  const view = vscode.window.createTreeView<ListNode>(options.id, { treeDataProvider: provider })
  view.title = options.title

  const reload = async (): Promise<void> => {
    const mine = ++generation
    stale = false
    await vscode.window.withProgress({ location: { viewId: options.id } }, async () => {
      let result: ListResult
      try {
        result = await options.load()
      } catch (error) {
        result = { nodes: [], message: `Could not load: ${error instanceof Error ? error.message : String(error)}` }
      }
      if (mine !== generation) return
      nodes = result.nodes
      view.message = result.message
      view.description = result.summary
      options.loaded?.(result)
      changed.fire(undefined)
    })
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
    invalidate,
    selected: () => view.selection[0],
    dispose:  () => {
      subscription.dispose()
      view.dispose()
      changed.dispose()
    },
  }
}
