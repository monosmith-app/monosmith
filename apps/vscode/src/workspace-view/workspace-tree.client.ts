import * as vscode from 'vscode'
import {
  countProjects, describeAffected, describeFilter, toTree,
  type TaskFacts, type TreeFolder, type TreeNode, type ViewFilter,
} from './workspace-tree.mapper.ts'

const VIEW_ID = 'monosmith.projects'
const DOT_SCHEME = 'monosmith-project'

/** The address a project row is decorated under: the dot of an affected project is its query. */
function dotUri (node: TreeNode): vscode.Uri {
  return vscode.Uri.from({ scheme: DOT_SCHEME, path: `/${encodeURIComponent(node.root)}/${node.project ?? ''}`, query: node.dot === true ? 'affected' : '' })
}

/** Whether a row is open: groups and folders are, a project when the filter found something under it, a task has nothing under it. */
function collapsibleState (node: TreeNode): vscode.TreeItemCollapsibleState {
  if (node.kind === 'task' || node.children.length === 0) return vscode.TreeItemCollapsibleState.None
  if (node.kind === 'project') return node.reveal === true ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.Collapsed

  return vscode.TreeItemCollapsibleState.Expanded
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
  /** How the view is narrowed now. */
  filter:            () => ViewFilter
  /** Narrows the view, or widens it again. */
  setFilter:         (filter: ViewFilter) => void
}

/**
 * Creates the view and adapts the tree to VS Code.
 * @param factsOf - what the last run did to a task id (`api:build`) in a folder
 * @param initial - how the view was narrowed last time
 */
export function createWorkspaceTree (factsOf: (root: string, task: string) => TaskFacts | undefined, initial?: ViewFilter): WorkspaceTree {
  const changed = new vscode.EventEmitter<TreeNode | undefined>()
  const decorations = new vscode.EventEmitter<vscode.Uri[]>()
  let folders: readonly TreeFolder[] = []
  let roots: TreeNode[] = []
  let narrowed: ViewFilter = initial ?? { text: '', affectedOnly: false }

  const provider: vscode.TreeDataProvider<TreeNode> = {
    onDidChangeTreeData: changed.event,
    getChildren:         (node) => (node === undefined ? roots : node.children),
    getTreeItem:         (node) => {
      const item = new vscode.TreeItem(node.label, collapsibleState(node))
      // A row the filter opens gets another id, so the view forgets that it was closed.
      item.id = node.reveal === true ? `${node.id}#open` : node.id
      item.description = node.description
      item.contextValue = node.kind === 'task' && node.hasLog === true ? 'task.log' : node.kind
      item.tooltip = node.tooltip
      item.accessibilityInformation = { label: node.spoken }
      item.iconPath = new vscode.ThemeIcon(node.icon)
      if (node.kind === 'project') {
        item.resourceUri = dotUri(node)
      }

      return item
    },
  }
  const view = vscode.window.createTreeView<TreeNode>(VIEW_ID, { treeDataProvider: provider, showCollapseAll: true })
  view.title = 'Projects'
  const decorator = vscode.window.registerFileDecorationProvider({
    onDidChangeFileDecorations: decorations.event,
    provideFileDecoration:      (uri) => (uri.scheme === DOT_SCHEME && uri.query === 'affected'
      ? { badge: '●', tooltip: 'Affected by the changes', color: new vscode.ThemeColor('charts.orange'), propagate: false }
      : undefined),
  })

  const dotUris = (nodes: readonly TreeNode[]): vscode.Uri[] => nodes.flatMap((node) => (node.kind === 'project' ? [dotUri(node)] : dotUris(node.children)))

  const rebuild = (): void => {
    const before = dotUris(roots)
    roots = toTree(folders, factsOf, narrowed)
    view.description = describeAffected(folders) || undefined
    view.message = describeFilter(narrowed, countProjects(roots))
    changed.fire(undefined)
    decorations.fire([...before, ...dotUris(roots)])
  }

  return {
    show:              (next) => { folders = next; rebuild() },
    redraw:            rebuild,
    selected:          () => view.selection[0],
    onSelectionChange: (handler) => view.onDidChangeSelection((event) => { handler(event.selection[0]) }),
    withProgress:      async (work) => await vscode.window.withProgress({ location: { viewId: VIEW_ID } }, work),
    filter:            () => narrowed,
    setFilter:         (next) => {
      narrowed = next
      rebuild()
    },
    dispose: () => {
      decorator.dispose()
      view.dispose()
      changed.dispose()
      decorations.dispose()
    },
  }
}
