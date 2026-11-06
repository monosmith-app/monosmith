import * as vscode from 'vscode'
import { checkProjectName } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, NeededTools, PlanAnswer } from './add-options.contract.ts'
import { tellDoctorFailure } from '../doctor-failure/index.ts'
import { languageRows, outcomeRows, whatIs, type OutcomeOption, type OutcomeRow } from './add-outcomes.mapper.ts'
import { choicesByAxis, roleOf } from './add-options.mapper.ts'
import { typedItem } from './add-choice-rows.mapper.ts'
import { pickChoices, type ChoicePrompts, type PicksCheck } from './pick-choices.use-case.ts'
import type { PicksAsked } from './check-picks.use-case.ts'
import { showNotice } from '../notice-buttons/index.ts'
import { askWebEnvironments, isWebApp } from '../web-environments/index.ts'
import { mainLanguage, startOnMainLanguage } from './main-language.mapper.ts'
import { canStartWithDefaults, defaultsRequest } from './defaults-request.mapper.ts'
import { say } from '../words/index.ts'

const ROLES: Record<string, string> = {
  app:     say('vscode.add-project.something-you-run-or-deploy-a'),
  lib:     say('vscode.add-project.code-the-other-projects-here-import'),
  package: say('vscode.add-project.published-to-a-registry-on-release'),
}

const TITLE = say('vscode.add-project.monosmith-add-project')

/**
 * Asks what to add with quick picks: what is being built (then the language where
 * the outcome has more than one), how to build it
 * (the default of the language, a stack, or choices such as a framework, a
 * test runner and a style) and the name.
 * @param options - what `smith kinds`, `smith choices` and `smith stacks` list
* @param existing - the names the workspace has already: a new project cannot take one
 * @param projects - the workspace's projects, to offer one to copy the shape of (`--like`)
 * @returns the request, `compose` when the person chose to compose a stack, or undefined when they pressed Escape at any question
 */
export async function askAddRequest (options: AddOptions, existing: readonly string[] = [], projects: ReadonlyArray<{ name: string, language: string }> = [], workspaceEnvironments: readonly string[] = [], check?: (asked: PicksAsked) => Promise<string>): Promise<AddRequest | 'compose' | undefined> {
  const found = await askOutcome(options, projects)
  if (found === undefined || found === 'compose') return found

  const { role, language, preset, outcomeId } = found
  const askName = async (title: string): Promise<string | undefined> => await vscode.window.showInputBox({
    title,
    prompt:        say('vscode.add-project.the-name-of-the-new', { role }),
    placeHolder:   say('vscode.add-project.for-example-billing-api'),
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? say('vscode.add-project.exists-already-pick-another-name', { value }) : undefined),
  })
  // The name comes first, then one question: the defaults (straight to the plan) or the questions one by one.
  let name: string | undefined
  if (canStartWithDefaults(found)) {
    name = await askName(`${TITLE}: name`)
    if (name === undefined) return undefined
    const setup = await askSetup()
    if (setup === undefined) return undefined
    if (setup === 'defaults') return defaultsRequest(found, name, workspaceEnvironments)
  }

  const build = await askBuild(options, language, preset, outcomeId, role, check)
  if (build === undefined) return undefined

  const like = await askLike(projects.filter((project) => project.language === language))
  if (like === null) return undefined

  name ??= await askName(`${TITLE} (4/4): name`)

  if (name === undefined) return undefined

  const request: AddRequest = { role, name, language, with: build.with, fresh: false, ...(build.stack !== undefined && { stack: build.stack }), ...(like !== undefined && { like }) }
  if (!isWebApp(request)) return request
  // A web app is built once per environment: smith refuses to add one that names none.
  const environments = await askWebEnvironments(name, workspaceEnvironments)

  return environments === undefined ? undefined : { ...request, environments }
}

/**
 * The question after the name: take the defaults and go to the plan, or answer each question.
 * The words are the ones of the terminal (docs/manual/tui.md).
 * @returns the answer; undefined when the person pressed Escape
 */
async function askSetup (): Promise<'defaults' | 'custom' | undefined> {
  const answer = await vscode.window.showQuickPick(
    [
      { label: say('guide.add_project_setup.start_with_the_defaults'), description: 'default', detail: say('vscode.add-project.the-answers-the-stack-suggests-what'), value: 'defaults' as const },
      { label: say('guide.add_project_setup.customise'), detail: say('vscode.add-project.answer-the-questions-one-at-a'), value: 'custom' as const },
    ],
    { title: say('vscode.add-project.how-do-you-want-to-set', { title: TITLE }), ignoreFocusOut: true },
  )

  return answer?.value
}

const EVERY_LANGUAGE = say('vscode.add-project.i-know-the-language-show-every')

/**
 * The first questions: what is being built, in plain words, then the language
 * only where the outcome leaves more than one. The old way (kind, then every
 * language) stays as the last line, and so does the composer.
 * @param options - what smith lists
 * @returns the kind, the language and the choice the outcome named; `compose` for the composer; undefined when the person pressed Escape
 */
async function askOutcome (options: AddOptions, projects: ReadonlyArray<{ language: string }>): Promise<'compose' | { role: string, language: string, preset: string[], outcomeId?: string } | undefined> {
  const found = await pickOutcomeRow(options.outcomes)
  if (found === undefined) return undefined
  const { row: picked, typed } = found
  if (picked.kind === 'compose') return 'compose'
  if (picked.kind === 'every') return askEveryLanguage(options, projects)
  if (picked.kind === 'none') return undefined

  const { outcome } = picked
  const language = outcome.languages.length === 1 ? outcome.languages[0] : await askLanguage(options, outcome.role, outcome.languages, projects, outcome)
  if (language === undefined) return undefined

  const preset = await askPreset(options, outcome, language, typed)

  return preset === undefined ? undefined : { role: outcome.role, language, preset, outcomeId: outcome.id }
}

/** The old way: the kind, then every language of that kind. */
async function askEveryLanguage (options: AddOptions, projects: ReadonlyArray<{ language: string }>): Promise<{ role: string, language: string, preset: string[], outcomeId?: string } | undefined> {
  const roles = [...new Set(options.kinds.map((kind) => roleOf(kind.kind, kind.language)))]
  const role = await vscode.window.showQuickPick(roles.map((name) => ({ label: name, description: ROLES[name] ?? '' })), { title: say('vscode.add-project.what-kind', { title: TITLE }), placeHolder: say('vscode.add-project.app-lib-or-package') })
  if (role === undefined) return undefined

  const language = await askLanguage(options, role.label, options.kinds.map((kind) => kind.language), projects)

  return language === undefined ? undefined : { role: role.label, language, preset: [] }
}

/**
 * Shows the first question: what is being built. The list is smith's own; the search
 * is done here, with the same rule as the terminal screen, so "react" finds a
 * website and "tauri" a desktop app. A search that finds nothing says so and names
 * the closest answers.
 * @param outcomes - what smith can add here
 * @returns the row picked and what was typed to find it; undefined when the person pressed Escape
 */
async function pickOutcomeRow (outcomes: readonly OutcomeOption[]): Promise<{ row: OutcomeRow, typed: string } | undefined> {
  const pick = vscode.window.createQuickPick<vscode.QuickPickItem & { row: OutcomeRow }>()
  pick.title = say('vscode.add-project.what-are-you-building', { title: TITLE })
  pick.ignoreFocusOut = true
  pick.placeholder = say('vscode.add-project.type-to-narrow-the-list-try')
  const show = (): void => {
    pick.items = outcomeRows(outcomes, pick.value).map((row) => ({ ...rowItem(row), row, alwaysShow: true }))
  }
  show()

  return await new Promise((resolve) => {
    pick.onDidChangeValue(show)
    pick.onDidAccept(() => {
      const row = pick.selectedItems[0]?.row
      // The "no result" row only explains: it keeps the list open.
      if (row === undefined || row.kind === 'none') return
      resolve({ row, typed: pick.value })
      pick.hide()
    })
    pick.onDidHide(() => {
      resolve(undefined)
      pick.dispose()
    })
    pick.show()
  })
}

function rowItem (row: OutcomeRow): vscode.QuickPickItem {
  switch (row.kind) {
    case 'outcome': {
      return { label: row.outcome.title, description: row.outcome.typical, detail: row.outcome.consequence }
    }
    case 'none': {
      return { label: say('vscode.add-project.no-result-for', { query: row.query }), description: row.closest.length === 0 ? '' : `closest: ${row.closest.join(' · ')}`, detail: say('vscode.add-project.change-the-words-or-use-one') }
    }
    case 'every': {
      return { label: EVERY_LANGUAGE, description: say('vscode.add-project.the-long-way-kind-first-then') }
    }
    case 'compose': {
      return { label: say('vscode.add-project.compose-a-stack'), description: say('vscode.add-project.search-frameworks-desktop-shells-tauri-electron') }
    }
  }
}

/**
 * Asks the language among those that fit, by the name a person knows, with
 * what each offers for the thing being built (the clouds of a function app).
 */
async function askLanguage (options: AddOptions, role: string, fitting: readonly string[], projects: ReadonlyArray<{ language: string }>, outcome?: OutcomeOption): Promise<string | undefined> {
  const main = mainLanguage(projects, fitting)
  const rows = startOnMainLanguage(languageRows(options.languages, options.kinds, role, fitting, outcome), main)
  const language = await vscode.window.showQuickPick(
    rows.map((row) => ({ label: row.label, description: row.language === main?.language ? `${row.description} ${main.why}`.trim() : row.description, detail: row.detail, language: row.language })),
    { title: say('vscode.add-project.which-language', { title: TITLE }), placeHolder: say(outcome?.items[0]?.layer === 'host' ? 'vscode.add-project.the-language-of-the-function-the' : 'vscode.add-project.pick-the-language'), matchOnDescription: true, matchOnDetail: true },
  )

  return language?.language
}

/**
 * Asks which framework, desktop shell or cloud the outcome uses; a website or a desktop app may decide later, a function app may not.
 * The list starts on the item the first question was searched for (`react` starts it on React), else on its first row.
 */
async function askPreset (options: AddOptions, outcome: OutcomeOption, language: string, typed: string): Promise<string[] | undefined> {
  const items = outcome.items.filter((item) => item.language === language)
  if (items.length === 0) return []

  const later = say('vscode.add-project.decide-in-the-next-question')
  const layer = outcome.items[0]?.layer ?? ''
  const question = options.layers[layer] === undefined ? say('guide.add_project_outcomes.which_one') : say('vscode.add-project.which', { toLowerCase: options.layers[layer].toLowerCase() })
  // "Decide in the next question" is first, so the list reads in order and Enter on it picks nothing.
  const laterRows = outcome.id === 'function' ? [] : [{ label: later, description: '', detail: say('vscode.add-project.you-can-still-pick-a-framework') }]
  const rows = [...laterRows, ...items.map((item) => ({ label: item.title ?? item.name, description: item.name, detail: item.description }))]
  const start = typedItem(items, typed)
  // Nothing typed: the cursor starts on "Decide in the next question", so Enter picks nothing.
  const startRow = rows.find((row) => (start === undefined ? row.label === later : row.description === start))
  const picked = await pickOne(rows, { title: `${TITLE}: ${question}`, placeHolder: say('vscode.add-project.your-pick-is-kept-the-next', { whatIs: whatIs(outcome.items[0]?.layer ?? '') }).trim() }, startRow)
  if (picked === undefined) return undefined

  return picked.label === later ? [] : [picked.description]
}

/**
 * Shows a list of one answer, with the cursor on a row the caller names (the
 * plain quick pick always starts on the first row).
 * @param rows - the lines of the list
 * @param shown - the title and the hint of the list
 * @param start - the row the cursor starts on; the first row when absent
 * @returns the row picked; undefined when the person pressed Escape
 */
async function pickOne<T extends vscode.QuickPickItem> (rows: readonly T[], shown: { title: string, placeHolder: string }, start?: T): Promise<T | undefined> {
  const pick = vscode.window.createQuickPick<T>()
  pick.title = shown.title
  pick.placeholder = shown.placeHolder
  pick.matchOnDescription = true
  pick.ignoreFocusOut = true
  pick.items = rows
  if (start !== undefined) pick.activeItems = [start]

  return await new Promise((resolve) => {
    let answer: T | undefined
    pick.onDidAccept(() => {
      answer = pick.selectedItems[0]
      pick.hide()
    })
    pick.onDidHide(() => {
      resolve(answer)
      pick.dispose()
    })
    pick.show()
  })
}

/**
 * Offers to copy the shape (build, test, layout, lint) of one project of the same language, when there are two or more.
 * @param siblings - the workspace's projects of the language
 * @returns the project to copy, undefined for what most siblings do, or null when the person pressed Escape
 */
async function askLike (siblings: ReadonlyArray<{ name: string, language: string }>): Promise<string | undefined | null> {
  if (siblings.length < 2) return undefined
  const majority = say('vscode.add-project.what-most-projects-do')
  const picked = await vscode.window.showQuickPick(
    [{ label: majority, description: say('vscode.add-project.smith-follows-the-siblings-that-agree') }, ...siblings.map((project) => ({ label: project.name, description: say('vscode.add-project.copy-the-shape-of-this-project') }))],
    { title: say('vscode.add-project.follow-which-project', { title: TITLE }), placeHolder: say('vscode.add-project.the-new-project-copies-the-build') },
  )
  if (picked === undefined) return null

  return picked.label === majority ? undefined : picked.label
}

/** The third question: the default of the language, a stack, or choices. */
async function askBuild (options: AddOptions, language: string, preset: readonly string[] = [], outcomeId?: string, role?: string, check?: (asked: PicksAsked) => Promise<string>): Promise<{ stack?: string, with: string[] } | undefined> {
  const stacks = options.stacks.filter((stack) => stack.language === language)
  const axes = choicesByAxis(options.choices, language, role)
  const items: Array<vscode.QuickPickItem & { kind_: 'default' | 'stack' | 'choices', stack?: string }> = [
    { label: say('vscode.add-project.the-default'), description: say('vscode.add-project.the-stack-the-workspace-uses-for', { language }), kind_: 'default' },
    ...stacks.map((stack) => ({ label: stack.name, description: stack.description, kind_: 'stack' as const, stack: stack.name })),
    ...(axes.length === 0 ? [] : [{ label: say('vscode.add-project.pick-choices'), description: say('vscode.add-project.a-framework-a-test-runner-a'), kind_: 'choices' as const }]),
  ]
  const how = preset.length > 0 ? items.find((item) => item.kind_ === 'choices') : (items.length === 1 ? items[0] : await vscode.window.showQuickPick(items, { title: say('vscode.add-project.3-4-how-is-it-built', { title: TITLE }), placeHolder: say('vscode.add-project.the-default-a-stack-or-your'), matchOnDescription: true }))
  if (how === undefined) return undefined
  if (how.kind_ === 'default') return { with: [] }
  if (how.kind_ === 'stack') return { stack: how.stack, with: [] }

  const picksCheck: PicksCheck | undefined = check === undefined ? undefined : async (names) => await check({ language, picks: names, ...(role !== undefined && { role }), ...(outcomeId !== undefined && { outcomeId }) })
  const picked = await pickChoices(options, language, preset, outcomeId, vscodeChoicePrompts, role, picksCheck)

  return picked === undefined ? undefined : { with: picked }
}

/** What Enter keeps in a single-choice list: the first row, the default, or none. */
function keptDefault (rows: Parameters<ChoicePrompts['one']>[1]): string {
  return rows[0]?.name === '' ? say('vscode.add-project.none-nothing-extra') : say('vscode.add-project.the-default-2', { part: rows[0]?.label ?? say('vscode.add-project.the-first-line') })
}

/** The questions of the choices, as quick picks. */
const vscodeChoicePrompts: ChoicePrompts = {
  many: async (question, rows) => {
    const defaults = rows.filter((row) => row.picked).map((row) => row.name)
    const answer = await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.name}`, placeHolder: `${question.hint} ${defaults.length > 0 ? say('vscode.add-project.enter-keeps-the-default', { join: defaults.join(', ') }) : say('vscode.add-project.tick-any-or-escape-to-cancel')}`.trim(), canPickMany: true, ignoreFocusOut: true })

    return answer?.map((row) => row.name)
  },
  // A single-choice axis is a radio list: one answer. Enter takes the first row, which is the default, or "None".
  one:   async (question, rows) => await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.name}`, placeHolder: say('vscode.add-project.enter-keeps', { hint: question.hint, part: keptDefault(rows) }).trim(), matchOnDetail: true }),
  again: async (problem) => await vscode.window.showWarningMessage(say('vscode.add-project.monosmith', { problem }), { modal: true }, say('vscode.add-project.pick-again')) === say('vscode.add-project.pick-again'),
  asked: async (question, rows, needing) => {
    const answer = await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.axis} for ${needing.join(', ')}`, placeHolder: say('vscode.add-project.smith-does-not-choose-the-for', { axis: question.axis }), matchOnDetail: true })

    return answer?.name
  },
}

/**
 * Shows the plan of `smith add --dry-run` and asks. A project that follows its
 * sibling projects can also be added with the defaults of its stack.
 * @param request - what the person asked for
 * @param plan - the text of the dry run
 * @param follows - whether the plan follows the siblings, so that `--fresh` would change it
 * @param choices - what the person chose, in words, and which of it is the default (`choicesInWords`); shown above the plan
 * @returns `add`, `fresh` (add with `--fresh`) or `cancel`
 */
export async function confirmAddPlan (request: AddRequest, plan: string, follows: boolean, choices = ''): Promise<PlanAnswer> {
  const add = say('vscode.add-project.add-project')
  const commitAdd = say('vscode.add-project.add-and-commit-so-undo-works')
  const fresh = say('vscode.add-project.add-with-the-stack-defaults')
  const commitFresh = say('vscode.add-project.stack-defaults-and-commit')
  const open = say('vscode.add-project.open-the-plan')
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      say('vscode.add-project.monosmith-add-the', { role: request.role, name: request.name, language: request.language }),
      { modal: true, detail: say('vscode.add-project.nothing-was-changed-yet-by-default', { part: choices === '' ? '' : `${choices}\n\n`, plan, part2: follows ? say('vscode.add-project.add-with-the-stack-defaults-skips') : '' }) },
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
      const undo = say('vscode.undo-change.undo')
      const commit = say('vscode.add-project.commit-now')
      const use = say('vscode.add-project.use-in-an-app')
      const library = outcome.role === 'lib'
      void vscode.window.showInformationMessage(
        say('vscode.add-project.monosmith-is-added-and-the-ci', { name: outcome.name, part: say(outcome.committed ? 'vscode.add-project.committed-as-one-change-undo-reverts' : 'vscode.add-project.not-committed-review-with-git-status'), part2: library ? say('vscode.add-project.use-it-from-an-app-pick') : '' }),
        ...(outcome.committed ? [undo] : (next === undefined ? [] : [commit])),
        ...(library ? [use] : []),
      ).then(async (choice) => {
        if (choice === undo) await vscode.commands.executeCommand('monosmith.undo')
        if (choice === commit) await next?.commit(outcome.operation)
        if (choice === use) await vscode.commands.executeCommand('monosmith.addDependency')
      })
      if (outcome.tools !== undefined) tellNeededTools(outcome.name, outcome.tools)
      break
    }
    case 'unchecked': {
      for (const line of outcome.lines) onOutput(line)
      void tellUnchecked(outcome, next)
      break
    }
    case 'failing': {
      tellDoctorFailure(outcome.failure, outcome.operation, outcome.lines, onOutput, next)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.add-project.monosmith-nothing-was-added', { message: outcome.message }))
      break
    }
    case 'composing':
    case 'cancelled': {
      break
    }
  }
}

/**
 * Says the new project needs tools the machine lacks, and offers the choice of
 * how to get them: the same `smith tools install` the Health view offers.
 * Nothing is installed before the person picks.
 */
function tellNeededTools (name: string, tools: NeededTools): void {
  const install = say('vscode.add-project.install-tools')
  const later = say('adopt.agent.decline')
  void showNotice(
    'warning',
    say('vscode.add-project.monosmith-needs-tools-this-machine-lacks', { name, join: tools.problems.join(' ') }),
    [
      { label: install, detail: say('vscode.add-project.choose-how-to-get-with-smith', { join: tools.names.join(', ') }) },
      { label: later, detail: say('vscode.add-project.go-on-smith-doctor-and-the') },
    ],
  ).then(async (choice) => {
    if (choice === install) await vscode.commands.executeCommand('monosmith.installTools')
  })
}

/** Says the project was made but could not be checked, and offers the ways on. Nothing is lost: the project stays on disk. */
async function tellUnchecked (outcome: Extract<AddOutcome, { kind: 'unchecked' }>, next?: AddFollowUp): Promise<void> {
  const open = say('vscode.add-project.install-page')
  const install = say('vscode.add-project.install-tools')
  const again = say('resultdialog.check.label')
  const skip = say('vscode.add-project.skip-the-check')
  const where = outcome.tool.dir === undefined ? '' : ` in ${outcome.tool.dir}`
  const files = outcome.tool.files > 0 ? ` (${outcome.tool.files} files)` : ''
  const check = outcome.tool.check
  for (;;) {
    const choice = await showNotice(
      'warning',
      say('vscode.add-project.monosmith-was-created-and-could-not', { name: outcome.name, where, files, title: outcome.tool.title }),
      [
        ...(outcome.tool.url === undefined ? [] : [{ label: open, detail: say('vscode.add-project.open-the-page-that-tells-how', { title: outcome.tool.title }) }]),
        { label: install, detail: say('vscode.add-project.install-the-missing-tools-with-smith') },
        ...(check === undefined || next === undefined ? [] : [{ label: again, detail: say('vscode.add-project.run-the-first-check-of-the') }]),
        { label: skip, detail: say('vscode.add-project.continue-without-the-check-the-health') },
      ],
    )
    if (choice === open && outcome.tool.url !== undefined) {
      await vscode.env.openExternal(vscode.Uri.parse(outcome.tool.url))
      continue
    }
    if (choice === install) await vscode.commands.executeCommand('monosmith.installTools')
    if (choice === again && check !== undefined) next?.checkAgain(check)
    if (choice === skip) void vscode.window.showInformationMessage(say('vscode.add-project.monosmith-not-checked-is-missing-the', { title: outcome.tool.title }))

    return
  }
}
