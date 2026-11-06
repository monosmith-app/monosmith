import * as vscode from 'vscode'
import type { ProjectActionOutcome, ProjectKind } from './project-action.contract.ts'
import { otherKinds } from './project-action-arguments.mapper.ts'
import { say } from '../words/index.ts'

/** A project, as far as these questions read it. */
export interface ProjectChoice {
  name:     string
  kind:     string
  language: string
  deps:     string[]
}

const WHAT_A_KIND_IS: Record<ProjectKind, string> = {
  app:     say('vscode.project-actions.something-you-run-or-deploy-other'),
  lib:     say('vscode.project-actions.code-the-other-projects-of-this'),
  package: say('vscode.project-actions.code-published-to-a-registry-on'),
}

/**
 * Asks which project, when the command was not started on one.
 * @param projects - the projects to choose from
 * @param title - what the command does, for the title of the list
 * @returns the project name, or undefined when the person cancels
 */
export async function pickProjectFor (projects: readonly ProjectChoice[], title: string): Promise<string | undefined> {
  const picked = await vscode.window.showQuickPick(projects.map((project) => ({ label: project.name, description: `${project.kind} · ${project.language}` })), { title: say('vscode.project-actions.monosmith', { title }), placeHolder: say('guide.actions.which_project') })

  return picked?.label
}

/**
 * Shows the plan of removing a project and asks the one thing smith asks: keep a copy of the folder, or delete it.
 * @param project - the project
 * @param plan - the text of `smith remove --dry-run`
 * @returns `keep-copy`, `delete`, or undefined when the person cancels
 */
export async function askRemoveChoice (project: string, plan: string): Promise<'keep-copy' | 'delete' | undefined> {
  const keep = say('vscode.project-actions.remove-it-and-keep-a-copy')
  const drop = say('vscode.project-actions.remove-it-and-delete-the-folder')
  const choice = await vscode.window.showWarningMessage(
    say('vscode.project-actions.monosmith-remove', { project }),
    { modal: true, detail: say('vscode.project-actions.keep-a-copy-moves-the-folder', { plan }) },
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
    { title: say('vscode.project-actions.monosmith-make-a', { project }), placeHolder: say('vscode.project-actions.is-now-what-should-it-become', { project, part: current === 'app' ? say('vscode.project-actions.an-app') : `a ${current}` }) },
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
  const go = say('vscode.project-actions.make-it', { part: kind === 'app' ? say('vscode.project-actions.an-app') : `a ${kind}` })
  const choice = await vscode.window.showInformationMessage(
    say('vscode.project-actions.monosmith-make', { project, part: kind === 'app' ? say('vscode.project-actions.an-app') : `a ${kind}` }),
    { modal: true, detail: say('vscode.project-actions.nothing-was-changed-yet-smith-moves', { plan }) },
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
  const picked = await vscode.window.showQuickPick(libraries.map((label) => ({ label })), { title: say('vscode.project-actions.monosmith-unlink-a-library-from', { project }), placeHolder: say('vscode.project-actions.which-library-should-the-project-stop') })

  return picked?.label
}

/**
 * Asks to unlink a library: the project stops depending on it, the library stays.
 * @param project - the project
 * @param library - the library
 * @returns true on a yes, undefined when the person cancels
 */
export async function confirmUnlink (project: string, library: string): Promise<true | undefined> {
  const go = say('vscode.project-actions.unlink-it')
  const choice = await vscode.window.showWarningMessage(
    say('vscode.project-actions.monosmith-unlink-from', { library, project }),
    { modal: true, detail: say('vscode.project-actions.stops-depending-on-the-library-stays', { project, library, project2: project }) },
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
      const commit = say('vscode.project-actions.commit-changes')
      void vscode.window.showInformationMessage(say('vscode.project-actions.monosmith-not-committed-review-with-git', { done }), commit).then(async (choice) => {
        if (choice === commit) await vscode.commands.executeCommand('monosmith.commit', { root })
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.convert-workspace.monosmith-nothing-was-changed', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
