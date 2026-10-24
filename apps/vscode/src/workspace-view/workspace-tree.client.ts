import * as vscode from 'vscode'
import type { TaskStatus } from '../task-run/index.ts'
import { toTree, type TreeFolder, type TreeNode } from './workspace-tree.mapper.ts'

const VIEW_ID = 'monosmith.projects'

const STATUS_ICONS: Record<TaskStatus, string> = {
  queued:  'circle-outline',
  running: 'sync~spin',
  passed:  'pass',
  failed:  'error',
  cached:  'history',
  skipped: 'debug-step-over',
}

/** The MonoSmith view of projects and their tasks. */
export interface WorkspaceTree extends vscode.Disposable {
  /** Shows the folders as they are now. */
  show:              (folders: readonly TreeFolder[]) => void
  /** Draws the rows again, for a changed task status. */
  redraw:            () => void
  /** The row that is selected, if there is one. */
  selected:          () => TreeNode | undefined
  /** Listens to the selection changing. */
  onSelectionChange: (handler: (node: TreeNode | undefined) => void) => vscode.Disposable
  /** Shows the progress bar of the view while the work goes. */
  withProgress:      <T>(work: () => Promise<T>) => Promise<T>
}

/**
 * Creates the view and adapts the tree to VS Code.
 * @param statusOf - the status of a task id (`api:build`) in a folder
 */
export function createWorkspaceTree (statusOf: (root: string, task: string) => TaskStatus | undefined): WorkspaceTree {
  const changed = new vscode.EventEmitter<TreeNode | undefined>()
  let folders: readonly TreeFolder[] = []
  let roots: TreeNode[] = []

  const rebuild = (): void => {
    roots = toTree(folders, statusOf)
    changed.fire(undefined)
  }

  const provider: vscode.TreeDataProvider<TreeNode> = {
    onDidChangeTreeData: changed.event,
    getChildren:         (node) => (node === undefined ? roots : node.children),
    getTreeItem:         (node) => {
      const collapsible = node.kind === 'task'
        ? vscode.TreeItemCollapsibleState.None
        : (node.kind === 'project' ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.Expanded)
      const item = new vscode.TreeItem(node.label, node.children.length === 0 && node.kind !== 'task' ? vscode.TreeItemCollapsibleState.None : collapsible)
      item.id = node.id
      item.description = node.description
      item.contextValue = node.kind
      item.tooltip = node.tooltip
      item.accessibilityInformation = { label: node.spoken }
      if (node.kind === 'task') {
        item.iconPath = new vscode.ThemeIcon(node.status === undefined ? 'play' : STATUS_ICONS[node.status])
      } else if (node.kind === 'project' && node.affected === true) {
        item.iconPath = new vscode.ThemeIcon('circle-filled', new vscode.ThemeColor('charts.orange'))
      }

      return item
    },
  }
  const view = vscode.window.createTreeView<TreeNode>(VIEW_ID, { treeDataProvider: provider, showCollapseAll: true })

  return {
    show:              (next) => { folders = next; rebuild() },
    redraw:            rebuild,
    selected:          () => view.selection[0],
    onSelectionChange: (handler) => view.onDidChangeSelection((event) => { handler(event.selection[0]) }),
    withProgress:      async (work) => await vscode.window.withProgress({ location: { viewId: VIEW_ID } }, work),
    dispose:           () => { view.dispose(); changed.dispose() },
  }
}
