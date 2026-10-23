import * as vscode from 'vscode'
import type { TaskStatus } from '../task-run/index.ts'
import type { WorkspaceSnapshot } from '../workspace/index.ts'
import { toTree, type TreeNode } from './workspace-tree.mapper.ts'

const STATUS_ICONS: Record<TaskStatus, string> = {
  queued:  'circle-outline',
  running: 'sync~spin',
  passed:  'pass',
  failed:  'error',
  cached:  'history',
  skipped: 'debug-step-over',
}

/** The MonoSmith view of projects and their tasks. */
export interface WorkspaceTree {
  provider: vscode.TreeDataProvider<TreeNode>
  /** Shows a new snapshot. */
  show:     (snapshot: WorkspaceSnapshot) => void
  /** Draws the rows again, for a changed task status. */
  redraw:   () => void
}

/**
 * Adapts the tree to VS Code.
 * @param statusOf - the status of a task id (`api:build`)
 */
export function createWorkspaceTree (statusOf: (task: string) => TaskStatus | undefined): WorkspaceTree {
  const changed = new vscode.EventEmitter<TreeNode | undefined>()
  let snapshot: WorkspaceSnapshot = { projects: [], affected: new Set() }
  let roots: TreeNode[] = []

  const rebuild = (): void => {
    roots = toTree(snapshot, statusOf)
    changed.fire(undefined)
  }

  const provider: vscode.TreeDataProvider<TreeNode> = {
    onDidChangeTreeData: changed.event,
    getChildren:         (node) => (node === undefined ? roots : node.children),
    getTreeItem:         (node) => {
      const item = new vscode.TreeItem(node.label, node.kind === 'task'
        ? vscode.TreeItemCollapsibleState.None
        : (node.kind === 'group' ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.Collapsed))
      item.id = node.id
      item.description = node.description
      item.contextValue = node.kind
      if (node.kind === 'task') {
        item.iconPath = new vscode.ThemeIcon(node.status === undefined ? 'play' : STATUS_ICONS[node.status])
      } else if (node.kind === 'project' && node.affected === true) {
        item.iconPath = new vscode.ThemeIcon('circle-filled', new vscode.ThemeColor('charts.orange'))
      }

      return item
    },
  }

  return {
    provider,
    show:   (next) => { snapshot = next; rebuild() },
    redraw: rebuild,
  }
}
