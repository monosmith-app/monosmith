import * as vscode from 'vscode'
import { checkProjectName } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, PlanAnswer } from './add-options.contract.ts'
import { tellDoctorFailure } from './doctor-failure-prompts.client.ts'
import { languageRows, outcomeRows, whatIs, type OutcomeOption, type OutcomeRow } from './add-outcomes.mapper.ts'
import { choicesByAxis, roleOf } from './add-options.mapper.ts'
import { typedItem } from './add-choice-rows.mapper.ts'
import { pickChoices, type ChoicePrompts } from './pick-choices.use-case.ts'
import { showNotice } from '../notice-buttons/index.ts'
import { isWebApp } from './web-environments.mapper.ts'
import { mainLanguage, startOnMainLanguage } from './main-language.mapper.ts'
import { askWebEnvironments } from './web-environments-prompts.client.ts'
import { canStartWithDefaults, defaultsRequest } from './defaults-request.mapper.ts'

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
 * @returns the request, `compose` when the person chose to compose a stack, or undefined when they pressed Escape at any question
 */
export async function askAddRequest (options: AddOptions, existing: readonly string[] = [], projects: ReadonlyArray<{ name: string, language: string }> = [], workspaceEnvironments: readonly string[] = []): Promise<AddRequest | 'compose' | undefined> {
  const found = await askOutcome(options, projects)
  if (found === undefined || found === 'compose') return found

  const { role, language, preset, outcomeId } = found
  const askName = async (title: string): Promise<string | undefined> => await vscode.window.showInputBox({
    title,
    prompt:        `The name of the new ${role}`,
    placeHolder:   'for example billing-api',
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? `${value} exists already: pick another name.` : undefined),
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

  const build = await askBuild(options, language, preset, outcomeId)
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
      { label: 'Start with the defaults', description: 'default', detail: 'The answers the stack suggests: what you picked, the shape of the other projects, and for a website lcl, dev, uat and prd with prd as the build. The plan shows every default as a flag.', value: 'defaults' as const },
      { label: 'Customise', detail: 'Answer the questions one at a time: choices, the shape of the other projects, the environments.', value: 'custom' as const },
    ],
    { title: `${TITLE}: how do you want to set it up?`, ignoreFocusOut: true },
  )

  return answer?.value
}

const EVERY_LANGUAGE = '$(list-flat) I know the language: show every language'

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
  const role = await vscode.window.showQuickPick(roles.map((name) => ({ label: name, description: ROLES[name] ?? '' })), { title: `${TITLE}: what kind?`, placeHolder: 'app, lib or package' })
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
  pick.title = `${TITLE}: what are you building?`
  pick.ignoreFocusOut = true
  pick.placeholder = 'Type to narrow the list: try react, azure function, electron, tauri, api'
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
      return { label: `$(info) No result for "${row.query}"`, description: row.closest.length === 0 ? '' : `closest: ${row.closest.join(' · ')}`, detail: 'Change the words, or use one of the two lines below.' }
    }
    case 'every': {
      return { label: EVERY_LANGUAGE, description: 'the long way: kind first, then the language from the full list' }
    }
    case 'compose': {
      return { label: '$(search) Compose a stack…', description: 'search frameworks, desktop shells (Tauri, Electron) and cloud hosts in one place' }
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
    { title: `${TITLE}: which language?`, placeHolder: outcome?.items[0]?.layer === 'host' ? 'The language of the function; the clouds it can run in are listed next to it' : 'Pick the language', matchOnDescription: true, matchOnDetail: true },
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

  const later = '$(clock) Decide in the next question'
  const layer = outcome.items[0]?.layer ?? ''
  const question = options.layers[layer] === undefined ? 'Which one?' : `Which ${options.layers[layer].toLowerCase()}?`
  // "Decide in the next question" is first, so the list reads in order and Enter on it picks nothing.
  const laterRows = outcome.id === 'function' ? [] : [{ label: later, description: '', detail: 'you can still pick a framework, a shell or a tool there' }]
  const rows = [...laterRows, ...items.map((item) => ({ label: item.title ?? item.name, description: item.name, detail: item.description }))]
  const start = typedItem(items, typed)
  // Nothing typed: the cursor starts on "Decide in the next question", so Enter picks nothing.
  const startRow = rows.find((row) => (start === undefined ? row.label === later : row.description === start))
  const picked = await pickOne(rows, { title: `${TITLE}: ${question}`, placeHolder: `${whatIs(outcome.items[0]?.layer ?? '')} Your pick is kept: the next questions do not ask it again`.trim() }, startRow)
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
  const majority = 'What most projects do'
  const picked = await vscode.window.showQuickPick(
    [{ label: majority, description: 'smith follows the siblings that agree' }, ...siblings.map((project) => ({ label: project.name, description: 'copy the shape of this project' }))],
    { title: `${TITLE}: follow which project?`, placeHolder: 'The new project copies the build, test, layout and lint of one project (smith add --like)' },
  )
  if (picked === undefined) return null

  return picked.label === majority ? undefined : picked.label
}

/** The third question: the default of the language, a stack, or choices. */
async function askBuild (options: AddOptions, language: string, preset: readonly string[] = [], outcomeId?: string): Promise<{ stack?: string, with: string[] } | undefined> {
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

  const picked = await pickChoices(options, language, preset, outcomeId, vscodeChoicePrompts)

  return picked === undefined ? undefined : { with: picked }
}

/** The questions of the choices, as quick picks. */
const vscodeChoicePrompts: ChoicePrompts = {
  many: async (question, rows) => {
    const defaults = rows.filter((row) => row.picked).map((row) => row.name)
    const answer = await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.name}`, placeHolder: `${question.hint} ${defaults.length > 0 ? `Enter keeps the default (${defaults.join(', ')})` : 'Tick any, or Escape to cancel'}`.trim(), canPickMany: true, ignoreFocusOut: true })

    return answer?.map((row) => row.name)
  },
  // A single-choice axis is a radio list: one answer. Enter takes the first row, which is the default, or "None".
  one:   async (question, rows) => await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.name}`, placeHolder: `${question.hint} Enter keeps ${rows[0]?.name === '' ? 'None (nothing extra)' : `${rows[0]?.label ?? 'the first line'} (the default)`}.`.trim(), matchOnDetail: true }),
  asked: async (question, rows, needing) => {
    const answer = await vscode.window.showQuickPick(rows, { title: `${TITLE}: ${question.axis} for ${needing.join(', ')}`, placeHolder: `smith does not choose the ${question.axis} for you: the first is suggested`, matchOnDetail: true })

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
  const add = 'Add project'
  const commitAdd = 'Add and commit (so Undo works)'
  const fresh = 'Add with the stack defaults'
  const commitFresh = 'Stack defaults, and commit'
  const open = 'Open the plan'
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      `MonoSmith: add the ${request.role} ${request.name} (${request.language})?`,
      { modal: true, detail: `${choices === '' ? '' : `${choices}\n\n`}${plan}\n\nNothing was changed yet. By default smith leaves the files uncommitted: review them with git. The commit buttons make it one commit that Undo can revert.${follows ? '\n\n"Add with the stack defaults" skips the shape of the sibling projects (smith add --fresh).' : ''}` },
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
      const commit = 'Commit now'
      const use = 'Use in an app…'
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
    case 'failing': {
      tellDoctorFailure(outcome.failure, outcome.operation, outcome.lines, onOutput, next)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was added. ${outcome.message}`)
      break
    }
    case 'composing':
    case 'cancelled': {
      break
    }
  }
}

/** Says the project was made but could not be checked, and offers the ways on. Nothing is lost: the project stays on disk. */
async function tellUnchecked (outcome: Extract<AddOutcome, { kind: 'unchecked' }>, next?: AddFollowUp): Promise<void> {
  const open = 'Install page'
  const install = 'Install tools'
  const again = 'Check again'
  const skip = 'Skip the check'
  const where = outcome.tool.dir === undefined ? '' : ` in ${outcome.tool.dir}`
  const files = outcome.tool.files > 0 ? ` (${outcome.tool.files} files)` : ''
  const check = outcome.tool.check
  for (;;) {
    const choice = await showNotice(
      'warning',
      `MonoSmith: ${outcome.name} was created${where}${files}, and could not be checked yet: ${outcome.tool.title} is not installed. Nothing is lost: your project stays on disk.`,
      [
        ...(outcome.tool.url === undefined ? [] : [{ label: open, detail: `Open the page that tells how to get ${outcome.tool.title}.` }]),
        { label: install, detail: 'Install the missing tools with smith tools install.' },
        ...(check === undefined || next === undefined ? [] : [{ label: again, detail: 'Run the first check of the new project again, in a terminal.' }]),
        { label: skip, detail: 'Continue without the check. The Health view keeps reminding you.' },
      ],
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
