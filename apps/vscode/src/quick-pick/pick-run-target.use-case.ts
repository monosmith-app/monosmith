import * as vscode from 'vscode'
import type { RunTarget } from '../task-run/index.ts'
import type { FolderWorkspace } from '../workspace/index.ts'
import { folderChoices, projectChoices, taskChoices, taskNameChoices } from './task-choices.mapper.ts'

/** What the person chose to run, and in which folder. */
export interface PickedTarget {
  root:   string
  target: RunTarget
}

/**
 * Asks which folder (when there are several), which project and which of its
 * tasks, or all of them.
 * @returns the target, or undefined when the person backed out
 */
export async function pickTask (folders: readonly FolderWorkspace[]): Promise<PickedTarget | undefined> {
  const folder = await pickFolder(folders)
  if (folder === undefined) return undefined
  const project = await vscode.window.showQuickPick(projectChoices(folder.snapshot), {
    title:              'Run Task: choose a project',
    placeHolder:        'Project',
    matchOnDescription: true,
    matchOnDetail:      true,
  })
  if (project === undefined) return undefined
  const task = await vscode.window.showQuickPick(taskChoices(project.value), {
    title:       `Run Task: ${project.value.name}`,
    placeHolder: 'Task',
  })
  if (task === undefined) return undefined

  return {
    root:   folder.root,
    target: task.value === undefined
      ? { kind: 'projects', projects: [project.value.name] }
      : { kind: 'task', project: project.value.name, task: task.value },
  }
}

/**
 * Asks which task to run across projects: in the affected ones, or in every
 * project that has it.
 * @returns the target, or undefined when the person backed out or nothing can run
 */
export async function pickTaskName (folders: readonly FolderWorkspace[], scope: 'affected' | 'all'): Promise<PickedTarget | undefined> {
  const folder = await pickFolder(folders)
  if (folder === undefined) return undefined
  const choices = taskNameChoices(folder.snapshot, scope === 'affected')
  if (choices.length === 0) {
    void vscode.window.showInformationMessage(scope === 'affected' ? 'MonoSmith: no project is affected by the changes.' : 'MonoSmith: no project has a task.')

    return undefined
  }
  const task = await vscode.window.showQuickPick(choices, {
    title:       scope === 'affected' ? 'Run Affected: choose a task' : 'Run in All Projects: choose a task',
    placeHolder: 'Task',
  })

  return task === undefined ? undefined : { root: folder.root, target: { kind: scope, task: task.value } }
}

async function pickFolder (folders: readonly FolderWorkspace[]): Promise<FolderWorkspace | undefined> {
  if (folders.length === 0) {
    void vscode.window.showInformationMessage('MonoSmith: there are no projects to run yet.')

    return undefined
  }
  if (folders.length === 1) return folders[0]
  const chosen = await vscode.window.showQuickPick(folderChoices(folders), { title: 'Choose a folder', placeHolder: 'Folder' })

  return folders.find((folder) => folder.root === chosen?.value)
}
