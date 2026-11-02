import * as vscode from 'vscode'
import { checkProjectName } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, ChoiceOption, PlanAnswer } from './add-options.contract.ts'
import { outcomesOf, type OutcomeOption } from './add-outcomes.mapper.ts'
import { askedAxes, choiceNames, choicesByAxis, hostTerms, roleOf } from './add-options.mapper.ts'

const ROLES: Record<string, string> = {
  app:     'something you run or deploy (a website, an API) → apps/<name>',
  lib:     'code the other projects here import → libs/<name>',
  package: 'published to a registry on release → packages/<name>',
}

const TITLE = 'MonoSmith: Add project'

/**
 * Asks what to add with quick picks: what is being built (then the language where
 * the outcome has more than one), how to build it
 * (the default of the language, a stack, or choices such as a framework, a
 * test runner and a style) and the name.
 * @param options - what `smith kinds`, `smith choices` and `smith stacks` list
* @param existing - the names the workspace has already: a new project cannot take one
 * @param projects - the workspace's projects, to offer one to copy the shape of (`--like`)
 * @returns the request, or undefined when the person pressed Escape at any question
 */
export async function askAddRequest (options: AddOptions, existing: readonly string[] = [], projects: ReadonlyArray<{ name: string, language: string }> = []): Promise<AddRequest | undefined> {
  const found = await askOutcome(options)
  if (found === undefined) return undefined

  const { role, language, preset } = found
  const build = await askBuild(options, language, preset)
  if (build === undefined) return undefined

  const like = await askLike(projects.filter((project) => project.language === language))
  if (like === null) return undefined

  const name = await vscode.window.showInputBox({
    title:         `${TITLE} (4/4): name`,
    prompt:        `The name of the new ${role}`,
    placeHolder:   'for example billing-api',
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? `${value} exists already: pick another name.` : undefined),
  })

  return name === undefined ? undefined : { role, name, language, with: build.with, fresh: false, ...(build.stack !== undefined && { stack: build.stack }), ...(like !== undefined && { like }) }
}

const EVERY_LANGUAGE = '$(list-flat) I know the language: show every language'

/**
 * The first questions: what is being built, in plain words, then the language
 * only where the outcome leaves more than one. The old way (kind, then every
 * language) stays as the last line, and so does the composer.
 * @param options - what smith lists
 * @returns the kind, the language and the choice the outcome named; undefined when the person pressed Escape or went to the composer
 */
async function askOutcome (options: AddOptions): Promise<{ role: string, language: string, preset: string[] } | undefined> {
  const compose = '$(search) Compose a stack…'
  const outcomes = outcomesOf(options)
  const picked = await vscode.window.showQuickPick(
    [
      ...outcomes.map((outcome) => ({ label: outcome.title, description: outcome.typical, detail: `${outcome.consequence}  (also found by: ${[...new Set(outcome.words)].slice(0, 8).join(', ')})`, outcome })),
      { label: EVERY_LANGUAGE, description: 'the long way: kind first, then the language from the full list', detail: '', outcome: undefined },
      { label: compose, description: 'search frameworks, desktop shells (Tauri, Electron) and cloud hosts in one place', detail: '', outcome: undefined },
    ],
    { title: `${TITLE}: what are you building?`, placeHolder: 'Type to narrow the list: try react, azure function, electron, tauri, api', matchOnDescription: true, matchOnDetail: true },
  )
  if (picked === undefined) return undefined
  if (picked.label === compose) {
    await vscode.commands.executeCommand('monosmith.composeStack')

    return undefined
  }
  if (picked.outcome === undefined) return askEveryLanguage(options)

  const { outcome } = picked
  const language = outcome.languages.length === 1 ? outcome.languages[0] : await askLanguage(options, outcome.role, outcome.languages)
  if (language === undefined) return undefined

  const preset = await askPreset(outcome, language)

  return preset === undefined ? undefined : { role: outcome.role, language, preset }
}

/** The old way: the kind, then every language of that kind. */
async function askEveryLanguage (options: AddOptions): Promise<{ role: string, language: string, preset: string[] } | undefined> {
  const roles = [...new Set(options.kinds.map((kind) => roleOf(kind.kind, kind.language)))]
  const role = await vscode.window.showQuickPick(roles.map((name) => ({ label: name, description: ROLES[name] ?? '' })), { title: `${TITLE}: what kind?`, placeHolder: 'app, lib or package' })
  if (role === undefined) return undefined

  const language = await askLanguage(options, role.label, options.kinds.map((kind) => kind.language))

  return language === undefined ? undefined : { role: role.label, language, preset: [] }
}

/** Asks the language among those that fit; the choices of a language are found by their names. */
async function askLanguage (options: AddOptions, role: string, fitting: readonly string[]): Promise<string | undefined> {
  const languages = options.kinds.filter((kind) => roleOf(kind.kind, kind.language) === role && fitting.includes(kind.language))
  const language = await vscode.window.showQuickPick(
    languages.map((kind) => ({ label: kind.language, description: kind.label, detail: [...choiceNames(options.choices, kind.language), ...hostTerms(options.choices, kind.language)].join(', ') })),
    { title: `${TITLE}: which language?`, placeHolder: 'A language, or a framework such as react', matchOnDescription: true, matchOnDetail: true },
  )

  return language?.label
}

/** Asks which framework, desktop shell or cloud the outcome uses; a website or a desktop app may decide later, a function app may not. */
async function askPreset (outcome: OutcomeOption, language: string): Promise<string[] | undefined> {
  const items = outcome.items.filter((item) => item.language === language)
  if (items.length === 0) return []

  const later = '$(clock) Decide in the next question'
  const question = { framework: 'Which framework?', shell: 'Which desktop shell?', host: 'Which cloud?' }[outcome.items[0]?.layer ?? ''] ?? 'Which one?'
  const picked = await vscode.window.showQuickPick(
    [...items.map((item) => ({ label: item.title ?? item.name, description: item.name, detail: item.description })), ...(outcome.id === 'function' ? [] : [{ label: later, description: '', detail: 'you can still pick a framework, a shell or a tool there' }])],
    { title: `${TITLE}: ${question}`, placeHolder: 'Your pick starts selected in the choices that follow', matchOnDescription: true },
  )
  if (picked === undefined) return undefined

  return picked.label === later ? [] : [picked.description]
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
async function askBuild (options: AddOptions, language: string, preset: readonly string[] = []): Promise<{ stack?: string, with: string[] } | undefined> {
  const stacks = options.stacks.filter((stack) => stack.language === language)
  const axes = choicesByAxis(options.choices, language)
  const items: Array<vscode.QuickPickItem & { kind_: 'default' | 'stack' | 'choices', stack?: string }> = [
    { label: '$(star) The default', description: `the stack the workspace uses for ${language}`, kind_: 'default' },
    ...stacks.map((stack) => ({ label: stack.name, description: stack.description, kind_: 'stack' as const, stack: stack.name })),
    ...(axes.length === 0 ? [] : [{ label: '$(list-selection) Pick choices…', description: 'a framework, a test runner, a style', kind_: 'choices' as const }]),
  ]
  const how = preset.length > 0 ? items.find((item) => item.kind_ === 'choices') : (items.length === 1 ? items[0] : await vscode.window.showQuickPick(items, { title: `${TITLE} (3/4): how is it built?`, placeHolder: 'The default, a stack, or your own choices', matchOnDescription: true }))
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
        picked:      choice.isDefault || preset.includes(choice.name),
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

/** What the notices of an add can do beyond showing text: the editor carries these out. */
export interface AddFollowUp {
  /** Commits the changes of the add as one operation, after its own confirmation. */
  commit:     (operation: string) => Promise<void>
  /** Runs the first check of the new project again, in a terminal. */
  checkAgain: (args: string[]) => void
}

/**
 * Says how adding ended, and offers what comes next: Undo after a commit, a
 * button that commits an add that was not committed, and the ways on when a
 * tool is missing. It does not wait for the person to read it.
 * @param outcome - what `addProject` returned
 * @param onOutput - gets every line smith printed, for the output channel
 * @param next - carries out the buttons that need the editor
 */
export function tellAddOutcome (outcome: AddOutcome, onOutput: (line: string) => void, next?: AddFollowUp): void {
  switch (outcome.kind) {
    case 'added': {
      for (const line of outcome.lines) onOutput(line)
      const undo = 'Undo'
      const commit = 'Commit these changes now'
      const use = 'Use it from an app…'
      const library = outcome.role === 'lib'
      void vscode.window.showInformationMessage(
        `MonoSmith: ${outcome.name} is added and the CI pipelines are regenerated. ${outcome.committed ? 'Committed as one change: Undo reverts it.' : 'Not committed: review with git status, or commit it now (one commit that Undo can revert).'}${library ? ' Use it from an app: pick the app, then this library.' : ''}`,
        ...(outcome.committed ? [undo] : (next === undefined ? [] : [commit])),
        ...(library ? [use] : []),
      ).then(async (choice) => {
        if (choice === undo) await vscode.commands.executeCommand('monosmith.undo')
        if (choice === commit) await next?.commit(outcome.operation)
        if (choice === use) await vscode.commands.executeCommand('monosmith.addDependency')
      })
      break
    }
    case 'unchecked': {
      for (const line of outcome.lines) onOutput(line)
      void tellUnchecked(outcome, next)
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

/** Says the project was made but could not be checked, and offers the ways on. Nothing is lost: the project stays on disk. */
async function tellUnchecked (outcome: Extract<AddOutcome, { kind: 'unchecked' }>, next?: AddFollowUp): Promise<void> {
  const open = 'Open the install page'
  const install = 'Install it with smith'
  const again = 'Check again'
  const skip = 'Continue without checking'
  const where = outcome.tool.dir === undefined ? '' : ` in ${outcome.tool.dir}`
  const files = outcome.tool.files > 0 ? ` (${outcome.tool.files} files)` : ''
  const check = outcome.tool.check
  for (;;) {
    const choice = await vscode.window.showWarningMessage(
      `MonoSmith: ${outcome.name} was created${where}${files}, and could not be checked yet: ${outcome.tool.title} is not installed. Nothing is lost: your project stays on disk.`,
      ...(outcome.tool.url === undefined ? [] : [open]),
      install,
      ...(check === undefined || next === undefined ? [] : [again]),
      skip,
    )
    if (choice === open && outcome.tool.url !== undefined) {
      await vscode.env.openExternal(vscode.Uri.parse(outcome.tool.url))
      continue
    }
    if (choice === install) await vscode.commands.executeCommand('monosmith.installTools')
    if (choice === again && check !== undefined) next?.checkAgain(check)
    if (choice === skip) void vscode.window.showInformationMessage(`MonoSmith: not checked. ${outcome.tool.title} is missing; the health view keeps reminding you.`)

    return
  }
}
