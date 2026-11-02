import * as vscode from 'vscode'
import type { ProjectActionOutcome, ProjectKind } from './project-action.contract.ts'
import { otherKinds } from './project-action-arguments.mapper.ts'

/** A project, as far as these questions read it. */
export interface ProjectChoice {
  name:     string
  kind:     string
  language: string
  deps:     string[]
}

const WHAT_A_KIND_IS: Record<ProjectKind, string> = {
  app:     'Something you run or deploy. Other projects do not import it.',
  lib:     'Code the other projects of this workspace import. It is not published.',
  package: 'Code published to a registry on release, for other people to install.',
}

/**
 * Asks which project, when the command was not started on one.
 * @param projects - the projects to choose from
 * @param title - what the command does, for the title of the list
 * @returns the project name, or undefined when the person cancels
 */
export async function pickProjectFor (projects: readonly ProjectChoice[], title: string): Promise<string | undefined> {
  const picked = await vscode.window.showQuickPick(projects.map((project) => ({ label: project.name, description: `${project.kind} · ${project.language}` })), { title: `MonoSmith: ${title}`, placeHolder: 'Which project?' })

  return picked?.label
}

/**
 * Shows the plan of removing a project and asks the one thing smith asks: keep a copy of the folder, or delete it.
 * @param project - the project
 * @param plan - the text of `smith remove --dry-run`
 * @returns `keep-copy`, `delete`, or undefined when the person cancels
 */
export async function askRemoveChoice (project: string, plan: string): Promise<'keep-copy' | 'delete' | undefined> {
  const keep = 'Remove it and keep a copy'
  const drop = 'Remove it and delete the folder'
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: remove ${project}?`,
    { modal: true, detail: `${plan}\n\n"Keep a copy" moves the folder to .smith/removed/ so you can get it back. "Delete" removes the folder for good, files that git cannot bring back included. Nothing is committed.` },
    keep,
    drop,
  )

  return choice === keep ? 'keep-copy' : (choice === drop ? 'delete' : undefined)
}

/**
 * Asks which kind a project becomes, with what each kind means.
 * @param project - the project
 * @param current - its kind now
 * @returns the kind, or undefined when the person cancels
 */
export async function askKind (project: string, current: string): Promise<ProjectKind | undefined> {
  const picked = await vscode.window.showQuickPick(
    otherKinds(current).map((kind) => ({ label: kind, detail: WHAT_A_KIND_IS[kind], value: kind })),
    { title: `MonoSmith: make ${project} a ...`, placeHolder: `${project} is ${current === 'app' ? 'an app' : `a ${current}`} now. What should it become?` },
  )

  return picked?.value
}

/**
 * Shows the plan of changing a kind: every file smith edits and the folder it moves.
 * @param project - the project
 * @param kind - the kind it becomes
 * @param plan - the text of `smith kind --dry-run`
 * @returns true on a yes, undefined when the person cancels
 */
export async function confirmKind (project: string, kind: ProjectKind, plan: string): Promise<true | undefined> {
  const go = `Make it ${kind === 'app' ? 'an app' : `a ${kind}`}`
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: make ${project} ${kind === 'app' ? 'an app' : `a ${kind}`}?`,
    { modal: true, detail: `${plan}\n\nNothing was changed yet. smith moves the folder and rewrites every place that names it. Nothing is committed.` },
    go,
  )

  return choice === go ? true : undefined
}

/**
 * Asks which linked library to unlink from a project.
 * @param project - the project
 * @param libraries - the libraries of this workspace that the project depends on
 * @returns the library, or undefined when the person cancels
 */
export async function askLinkedLibrary (project: string, libraries: readonly string[]): Promise<string | undefined> {
  const picked = await vscode.window.showQuickPick(libraries.map((label) => ({ label })), { title: `MonoSmith: unlink a library from ${project}`, placeHolder: 'Which library should the project stop using?' })

  return picked?.label
}

/**
 * Asks to unlink a library: the project stops depending on it, the library stays.
 * @param project - the project
 * @param library - the library
 * @returns true on a yes, undefined when the person cancels
 */
export async function confirmUnlink (project: string, library: string): Promise<true | undefined> {
  const go = 'Unlink it'
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: unlink ${library} from ${project}?`,
    { modal: true, detail: `${project} stops depending on ${library}. The library stays in this workspace, and other projects keep using it. smith edits the manifest of ${project} and runs the install. Nothing is committed.` },
    go,
  )

  return choice === go ? true : undefined
}

/**
 * Says how an action on a project ended, and offers the Commit command after a success
 * (smith leaves the change uncommitted). It does not wait for the person to read it.
 * @param outcome - what `planThenApply` returned
 * @param done - the sentence for a success, such as "web is removed"
 * @param onOutput - gets every line smith printed, for the output channel
 * @param root - the workspace folder, so the Commit command works on the right one
 */
export function tellActionOutcome (outcome: ProjectActionOutcome, done: string, onOutput: (line: string) => void, root: string): void {
  switch (outcome.kind) {
    case 'done': {
      for (const line of outcome.lines) onOutput(line)
      const commit = 'Commit changes…'
      void vscode.window.showInformationMessage(`MonoSmith: ${done}. Not committed: review with git status, or commit it so Undo can revert it.`, commit).then(async (choice) => {
        if (choice === commit) await vscode.commands.executeCommand('monosmith.commit', { root })
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was changed. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
