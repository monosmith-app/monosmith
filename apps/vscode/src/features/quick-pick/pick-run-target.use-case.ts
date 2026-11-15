import * as vscode from 'vscode'
import type { RunTarget } from '../task-run/index.ts'
import type { FolderWorkspace } from '../workspace/index.ts'
import { everyProjectChoice, folderChoices, projectChoices, taskChoices, taskNameChoices } from './task-choices.mapper.ts'
import { say } from '../words/index.ts'

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
    title:              say('vscode.quick-pick.run-task-choose-a-project'),
    placeHolder:        say('vscode.quick-pick.project'),
    matchOnDescription: true,
    matchOnDetail:      true,
  })
  if (project === undefined) return undefined
  const task = await vscode.window.showQuickPick(taskChoices(project.value), {
    title:       say('vscode.quick-pick.run-task', { name: project.value.name }),
    placeHolder: say('vscode.quick-pick.task'),
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
    void vscode.window.showInformationMessage(say(scope === 'affected' ? 'vscode.quick-pick.monosmith-no-project-is-affected-by' : 'vscode.quick-pick.monosmith-no-project-has-a-task'))

    return undefined
  }
  const task = await vscode.window.showQuickPick(choices, {
    title:       say(scope === 'affected' ? 'vscode.quick-pick.run-affected-choose-a-task' : 'vscode.quick-pick.run-in-all-projects-choose-a'),
    placeHolder: say('vscode.quick-pick.task'),
  })

  return task === undefined ? undefined : { root: folder.root, target: { kind: scope, task: task.value } }
}

async function pickFolder (folders: readonly FolderWorkspace[]): Promise<FolderWorkspace | undefined> {
  if (folders.length === 0) {
    void vscode.window.showInformationMessage(say('vscode.quick-pick.monosmith-there-are-no-projects-to'))

    return undefined
  }
  if (folders.length === 1) return folders[0]
  const chosen = await vscode.window.showQuickPick(folderChoices(folders), { title: say('vscode.command-host.choose-a-folder'), placeHolder: say('vscode.command-host.folder') })

  return folders.find((folder) => folder.root === chosen?.value)
}

/**
 * Asks which project (and which folder, when there are several) an action is
 * about, such as explaining or drawing its graph.
 * @param title - what the action is, for the title of the pick
 * @returns the folder and the project, or undefined when the person backed out
 */
export async function pickProject (folders: readonly FolderWorkspace[], title: string): Promise<{ root: string, project: string } | undefined> {
  const folder = await pickFolder(folders)
  if (folder === undefined) return undefined
  const project = await vscode.window.showQuickPick(everyProjectChoice(folder.snapshot), { title, placeHolder: say('vscode.quick-pick.project'), matchOnDescription: true })

  return project === undefined ? undefined : { root: folder.root, project: project.value }
}
