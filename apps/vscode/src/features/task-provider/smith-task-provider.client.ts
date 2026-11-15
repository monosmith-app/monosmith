import * as vscode from 'vscode'
import { toOfferedTasks } from './offered-tasks.mapper.ts'
import { nameTask, toSmithArguments } from './smith-arguments.mapper.ts'
import type { SmithTaskDefinition } from './smith-task-definition.contract.ts'

/** What the provider needs to know about a folder that runs smith. */
export interface TaskFolder {
  root:       string
  /** The smith the folder runs. */
  executable: string
  /** The task names of its projects. */
  tasks:      string[]
}

/** The task type of `tasks.json`. */
export const TASK_TYPE = 'smith'

/**
 * Offers smith's tasks to VS Code: Run Task lists them, and `tasks.json` can
 * name them with `"type": "smith"`. They run `smith run` in the terminal, and
 * the `$smith` problem matcher reads the failures out of the output.
 * @param folders - the folders that run smith now
 */
export function registerSmithTasks (folders: () => readonly TaskFolder[]): vscode.Disposable {
  const build = (definition: SmithTaskDefinition, folder: TaskFolder, scope: vscode.WorkspaceFolder, name: string, group?: vscode.TaskGroup): vscode.Task | undefined => {
    const args = toSmithArguments(definition)
    if (args === undefined) return undefined
    const task = new vscode.Task(
      definition,
      scope,
      name,
      TASK_TYPE,
      new vscode.ShellExecution({ value: folder.executable, quoting: vscode.ShellQuoting.Strong }, args, { cwd: folder.root }),
      ['$smith', '$smith-tsc', '$smith-eslint'],
    )
    task.group = group
    task.detail = `smith ${args.join(' ')}`

    return task
  }
  const scopeOf = (root: string): vscode.WorkspaceFolder | undefined => vscode.workspace.workspaceFolders?.find((folder) => folder.uri.fsPath === root)
  const groupOf = (name: string | undefined): vscode.TaskGroup | undefined => {
    switch (name) {
      case 'build': { return vscode.TaskGroup.Build }
      case 'test': { return vscode.TaskGroup.Test }
      case 'clean': { return vscode.TaskGroup.Clean }
      default: { return undefined }
    }
  }

  return vscode.tasks.registerTaskProvider(TASK_TYPE, {
    provideTasks: () => folders().flatMap((folder) => {
      const scope = scopeOf(folder.root)

      return scope === undefined ? [] : toOfferedTasks(folder.tasks).flatMap((offered) => build(offered.definition, folder, scope, offered.name, groupOf(offered.group)) ?? [])
    }),
    resolveTask: (task) => {
      const definition = task.definition as unknown as SmithTaskDefinition
      const scope = typeof task.scope === 'object' ? task.scope : vscode.workspace.workspaceFolders?.[0]
      const folder = scope === undefined ? undefined : folders().find((candidate) => candidate.root === scope.uri.fsPath)

      return scope === undefined || folder === undefined ? undefined : build(definition, folder, scope, nameTask(definition))
    },
  })
}
