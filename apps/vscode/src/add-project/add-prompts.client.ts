import * as vscode from 'vscode'
import { checkProjectName } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, PlanAnswer } from './add-options.contract.ts'
import { choicesByAxis, roleOf } from './add-options.mapper.ts'

const ROLES: Record<string, string> = {
  app:     'something you run or deploy (a website, an API) → apps/<name>',
  lib:     'code the other projects here import → libs/<name>',
  package: 'published to a registry on release → packages/<name>',
}

const TITLE = 'MonoSmith: Add project'

/**
 * Asks what to add with quick picks: the kind, the language, how to build it
 * (the default of the language, a stack, or choices such as a framework, a
 * test runner and a style) and the name.
 * @param options - what `smith kinds`, `smith choices` and `smith stacks` list
 * @returns the request, or undefined when the person pressed Escape at any question
 */
export async function askAddRequest (options: AddOptions): Promise<AddRequest | undefined> {
  const roles = [...new Set(options.kinds.map((kind) => roleOf(kind.kind, kind.language)))]
  const role = await vscode.window.showQuickPick(roles.map((name) => ({ label: name, description: ROLES[name] ?? '' })), { title: `${TITLE} (1/4): what kind?`, placeHolder: 'app, lib or package' })
  if (role === undefined) return undefined

  const languages = options.kinds.filter((kind) => roleOf(kind.kind, kind.language) === role.label)
  const language = await vscode.window.showQuickPick(
    languages.map((kind) => ({ label: kind.language, description: kind.label })),
    { title: `${TITLE} (2/4): which language?`, placeHolder: 'The language of the project', matchOnDescription: true },
  )
  if (language === undefined) return undefined

  const build = await askBuild(options, language.label)
  if (build === undefined) return undefined

  const name = await vscode.window.showInputBox({
    title:         `${TITLE} (4/4): name`,
    prompt:        `The name of the new ${role.label}`,
    placeHolder:   'billing-api',
    validateInput: (value) => checkProjectName(value),
  })

  return name === undefined ? undefined : { role: role.label, name, language: language.label, with: build.with, fresh: false, ...(build.stack !== undefined && { stack: build.stack }) }
}

/** The third question: the default of the language, a stack, or choices. */
async function askBuild (options: AddOptions, language: string): Promise<{ stack?: string, with: string[] } | undefined> {
  const stacks = options.stacks.filter((stack) => stack.language === language)
  const axes = choicesByAxis(options.choices, language)
  const items: Array<vscode.QuickPickItem & { kind_: 'default' | 'stack' | 'choices', stack?: string }> = [
    { label: '$(star) The default', description: `the stack the workspace uses for ${language}`, kind_: 'default' },
    ...stacks.map((stack) => ({ label: stack.name, description: stack.description, kind_: 'stack' as const, stack: stack.name })),
    ...(axes.length === 0 ? [] : [{ label: '$(list-selection) Pick choices…', description: 'a framework, a test runner, a style', kind_: 'choices' as const }]),
  ]
  const how = items.length === 1 ? items[0] : await vscode.window.showQuickPick(items, { title: `${TITLE} (3/4): how is it built?`, placeHolder: 'The default, a stack, or your own choices', matchOnDescription: true })
  if (how === undefined) return undefined
  if (how.kind_ === 'default') return { with: [] }
  if (how.kind_ === 'stack') return { stack: how.stack, with: [] }

  const picked: string[] = []
  for (const axis of axes) {
    const defaults = axis.choices.filter((choice) => choice.isDefault).map((choice) => choice.name)
    const answer = await vscode.window.showQuickPick(
      axis.choices.map((choice) => ({
        label:       choice.name,
        description: `${choice.isDefault ? '(default) ' : ''}${choice.requires.length > 0 ? `needs ${choice.requires.join(', ')}` : ''}`.trim(),
        detail:      choice.description,
        picked:      choice.isDefault,
      })),
      { title: `${TITLE}: ${axis.axis}`, placeHolder: defaults.length > 0 ? `Enter keeps the default (${defaults.join(', ')})` : `Choose ${axis.axis}, or Escape to cancel`, canPickMany: true },
    )
    if (answer === undefined) return undefined
    if (!axis.many && answer.length > 1) {
      void vscode.window.showWarningMessage(`MonoSmith: ${axis.axis} takes one choice; smith will refuse more. Pick one and run Add project again.`)

      return undefined
    }
    picked.push(...answer.map((item) => item.label))
  }

  return { with: picked }
}

/**
 * Shows the plan of `smith add --dry-run` and asks. A project that follows its
 * sibling projects can also be added with the defaults of its stack.
 * @param request - what the person asked for
 * @param plan - the text of the dry run
 * @param follows - whether the plan follows the siblings, so that `--fresh` would change it
 * @returns `add`, `fresh` (add with `--fresh`) or `cancel`
 */
export async function confirmAddPlan (request: AddRequest, plan: string, follows: boolean): Promise<PlanAnswer> {
  const add = 'Add project'
  const fresh = 'Add with the stack defaults'
  const open = 'Open the plan'
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      `MonoSmith: add the ${request.role} ${request.name} (${request.language})?`,
      { modal: true, detail: `${plan}\n\nNothing was changed yet. smith leaves the files uncommitted; review them with git, or undo the change.${follows ? '\n\n"Add with the stack defaults" skips the shape of the sibling projects (smith add --fresh).' : ''}` },
      add,
      ...(follows ? [fresh] : []),
      open,
    )
    if (choice === open) {
      const document = await vscode.workspace.openTextDocument({ language: 'plaintext', content: `${plan}\n` })
      await vscode.window.showTextDocument(document, { preview: true })
      continue
    }
    if (choice === add) return 'add'

    return choice === fresh ? 'fresh' : 'cancel'
  }
}

/**
 * Says how adding ended, and offers Undo after a success. It does not wait for the person to read it.
 * @param outcome - what `addProject` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellAddOutcome (outcome: AddOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'added': {
      for (const line of outcome.lines) onOutput(line)
      const undo = 'Undo'
      void vscode.window.showInformationMessage(
        `MonoSmith: ${outcome.name} is added and the CI pipelines are regenerated. The files are not committed: review them with git.`,
        undo,
      ).then(async (choice) => {
        if (choice === undo) await vscode.commands.executeCommand('monosmith.undo')
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was added. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
