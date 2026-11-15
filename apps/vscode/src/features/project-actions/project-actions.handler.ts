import * as vscode from 'vscode'
import { askCommitBox } from '../commit-box/index.ts'
import { projectGiven, type CommandHost } from '../command-host/index.ts'
import { planThenApply } from './plan-then-apply.use-case.ts'
import { kindArguments, kindPlanArguments, removeArguments, removePlanArguments, unlinkArguments } from './project-action-arguments.mapper.ts'
import { askKind, askLinkedLibrary, askRemoveChoice, confirmKind, confirmUnlink, pickProjectFor, tellActionOutcome } from './project-action-prompts.client.ts'
import { say } from '../words/index.ts'

/**
 * Runs an action on a project from its menu or the palette: the project is the row or a pick,
 * smith shows a plan, and the change runs only on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 * @param action - what to do with the project
 */
export async function projectAction (host: CommandHost, argument: unknown, action: 'remove' | 'kind' | 'unlink'): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite(say('vscode.project-actions.changing-a-project'))) return
  const projects = state.snapshot?.projects ?? []
  const title = { remove: say('vscode.project-actions.remove-a-project'), kind: say('vscode.project-actions.make-a-project-another-kind'), unlink: say('vscode.project-actions.unlink-a-library') }[action]
  const name = projectGiven(argument) ?? await pickProjectFor(projects, title)
  const project = projects.find((item) => item.name === name)
  if (project === undefined) return
  host.showOutput()
  const run = host.smithIn(state)
  const apply = async (args: string[]): ReturnType<typeof run> => await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: say('vscode.project-actions.monosmith-in', { title, name: state.name }) },
    async () => await host.smithIn(state, true, 600_000)(args),
  )
  let outcome
  let done: string
  if (action === 'remove') {
    outcome = await planThenApply({ run, apply, confirm: async (plan) => await askRemoveChoice(project.name, plan) }, removePlanArguments(project.name), (keep) => removeArguments(project.name, keep))
    done = `${project.name} is removed`
  } else if (action === 'kind') {
    const kind = await askKind(project.name, project.kind)
    if (kind === undefined) return
    outcome = await planThenApply({ run, apply, confirm: async (plan) => await confirmKind(project.name, kind, plan) }, kindPlanArguments(project.name, kind), () => kindArguments(project.name, kind))
    done = `${project.name} is now ${kind === 'app' ? say('vscode.project-actions.an-app') : `a ${kind}`}`
  } else {
    const linked = project.deps.filter((dep) => projects.some((item) => item.name === dep && item.kind !== 'app'))
    if (linked.length === 0) {
      void vscode.window.showInformationMessage(say('vscode.project-actions.monosmith-uses-no-library-of-this', { name: project.name }))

      return
    }
    const library = await askLinkedLibrary(project.name, linked)
    if (library === undefined) return
    outcome = await planThenApply(
      {
        run,
        apply,
        confirm: async () => {
          const commit = await askCommitBox(say('vscode.project-actions.unlink-a-library'))

          return commit !== undefined && await confirmUnlink(project.name, library) ? { commit } : undefined
        },
      },
      undefined,
      (answer) => [...unlinkArguments(project.name, library), ...(answer.commit ? ['--commit'] : [])],
    )
    done = say('vscode.project-actions.is-unlinked-from', { library, name: project.name })
  }
  host.report(`${action} ${project.name}`, state, outcome)
  tellActionOutcome(outcome, done, host.note, state.root)
  if (outcome.kind === 'done') await host.refreshFolder(state.root, 'full', false)
}
