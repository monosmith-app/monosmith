import * as vscode from 'vscode'
import { checkProjectName } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, ChoiceOption, PlanAnswer } from './add-options.contract.ts'
import { askedAxes, choiceNames, choicesByAxis, hostTerms, roleOf } from './add-options.mapper.ts'

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
* @param existing - the names the workspace has already: a new project cannot take one
 * @param projects - the workspace's projects, to offer one to copy the shape of (`--like`)
 * @returns the request, or undefined when the person pressed Escape at any question
 */
export async function askAddRequest (options: AddOptions, existing: readonly string[] = [], projects: ReadonlyArray<{ name: string, language: string }> = []): Promise<AddRequest | undefined> {
  const roles = [...new Set(options.kinds.map((kind) => roleOf(kind.kind, kind.language)))]
  const role = await vscode.window.showQuickPick(roles.map((name) => ({ label: name, description: ROLES[name] ?? '' })), { title: `${TITLE} (1/4): what kind?`, placeHolder: 'app, lib or package' })
  if (role === undefined) return undefined

  const languages = options.kinds.filter((kind) => roleOf(kind.kind, kind.language) === role.label)
  const language = await vscode.window.showQuickPick(
    languages.map((kind) => ({ label: kind.language, description: kind.label, detail: [...choiceNames(options.choices, kind.language), ...hostTerms(options.choices, kind.language)].join(', ') })),
    { title: `${TITLE} (2/4): which language?`, placeHolder: 'A language, or a framework such as react', matchOnDescription: true, matchOnDetail: true },
  )
  if (language === undefined) return undefined

  const build = await askBuild(options, language.label)
  if (build === undefined) return undefined

  const like = await askLike(projects.filter((project) => project.language === language.label))
  if (like === null) return undefined

  const name = await vscode.window.showInputBox({
    title:         `${TITLE} (4/4): name`,
    prompt:        `The name of the new ${role.label}`,
    placeHolder:   'for example billing-api',
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? `${value} exists already: pick another name.` : undefined),
  })

  return name === undefined ? undefined : { role: role.label, name, language: language.label, with: build.with, fresh: false, ...(build.stack !== undefined && { stack: build.stack }), ...(like !== undefined && { like }) }
}

/**
 * Offers to copy the shape (build, test, layout, lint) of one project of the same language, when there are two or more.
 * @param siblings - the workspace's projects of the language
 * @returns the project to copy, undefined for what most siblings do, or null when the person pressed Escape
 */
async function askLike (siblings: ReadonlyArray<{ name: string, language: string }>): Promise<string | undefined | null> {
  if (siblings.length < 2) return undefined
  const majority = 'What most projects do'
  const picked = await vscode.window.showQuickPick(
    [{ label: majority, description: 'smith follows the siblings that agree' }, ...siblings.map((project) => ({ label: project.name, description: 'copy the shape of this project' }))],
    { title: `${TITLE}: follow which project?`, placeHolder: 'The new project copies the build, test, layout and lint of one project (smith add --like)' },
  )
  if (picked === undefined) return null

  return picked.label === majority ? undefined : picked.label
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
  const asked = askedAxes(options.choices, language)
  for (const axis of axes) {
    if (asked.includes(axis.axis)) {
      const answer = await askOneOf(axis, options.choices.filter((choice) => picked.includes(choice.name)))
      if (answer === null) return undefined
      if (answer !== undefined) picked.push(answer)
      continue
    }
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
 * Asks for the choice of an axis that smith never picks (a trigger, a runtime):
 * one answer, the suggested one first, left out when no picked choice needs the axis.
 * @param axis - the axis and its choices
 * @param picked - the choices picked so far
 * @returns the name picked, undefined when no picked choice needs the axis, or null when the person pressed Escape
 */
async function askOneOf (axis: { axis: string, choices: ChoiceOption[] }, picked: readonly ChoiceOption[]): Promise<string | undefined | null> {
  const needing = picked.filter((choice) => choice.needs.includes(axis.axis))
  if (needing.length === 0) return undefined
  const fits = axis.choices
    .filter((choice) => picked.every((other) => !choice.conflicts.includes(other.name) && !other.conflicts.includes(choice.name)))
    .sort((a, b) => Number(b.suggested) - Number(a.suggested))
  const answer = await vscode.window.showQuickPick(
    fits.map((choice) => ({ label: choice.name, description: choice.suggested ? '(suggested)' : '', detail: choice.description })),
    { title: `${TITLE}: ${axis.axis} for ${needing.map((choice) => choice.name).join(', ')}`, placeHolder: `smith does not choose the ${axis.axis} for you: the first is suggested`, matchOnDetail: true },
  )

  return answer === undefined ? null : answer.label
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
  const commitAdd = 'Add and commit (so Undo works)'
  const fresh = 'Add with the stack defaults'
  const commitFresh = 'Stack defaults, and commit'
  const open = 'Open the plan'
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      `MonoSmith: add the ${request.role} ${request.name} (${request.language})?`,
      { modal: true, detail: `${plan}\n\nNothing was changed yet. By default smith leaves the files uncommitted: review them with git. The commit buttons make it one commit that Undo can revert.${follows ? '\n\n"Add with the stack defaults" skips the shape of the sibling projects (smith add --fresh).' : ''}` },
      add,
      commitAdd,
      ...(follows ? [fresh, commitFresh] : []),
      open,
    )
    if (choice === open) {
      const document = await vscode.workspace.openTextDocument({ language: 'plaintext', content: `${plan}\n` })
      await vscode.window.showTextDocument(document, { preview: true })
      continue
    }
    if (choice === add) return 'add'
    if (choice === commitAdd) return 'add-commit'
    if (choice === fresh) return 'fresh'

    return choice === commitFresh ? 'fresh-commit' : 'cancel'
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
        `MonoSmith: ${outcome.name} is added and the CI pipelines are regenerated. ${outcome.committed ? 'Committed as one change: Undo reverts it.' : 'Not committed: review with git status.'}`,
        ...(outcome.committed ? [undo] : []),
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
