import { describeFailure, type SmithRun } from '../../shared/smith-run/index.ts'
import type { GuidedAnswers, GuidedCommand, GuidedOutcome, GuidedStep, ResolvedStep } from './guided-command.contract.ts'
import { environmentsOfApp, parseArchitectures, parseBuiltinPlugins, parseMarketplace, parsePipelineTemplates, parseProjectNames, parseProjectView, parseSettings, parseWebApps, parseWorkspaceEnvironments, settingWords, type ProjectView, type Setting, type WebApp } from './guided-sources.mapper.ts'
import { flag } from '../../shared/smith-flags/index.ts'
import { say } from '../../shared/words/index.ts'

/** What running a guided command needs from the editor. */
export interface GuidedDeps {
  /** Runs smith in the workspace folder. */
  run:        (args: string[]) => Promise<SmithRun>
  /** Asks one step. Resolves the answer (a list for `multi`), or undefined when the person cancels. */
  ask:        (step: ResolvedStep) => Promise<string | string[] | undefined>
  /** Shows the text of a `read` command. */
  show:       (title: string, text: string) => Promise<void>
  /** Shows what will change and asks. Nothing is changed before a yes. */
  confirm:    (title: string, detail: string, verb: string) => Promise<boolean>
  /** Asks the unticked commit box, for a command that takes `--commit`. Resolves undefined when the person cancels. */
  askCommit?: () => Promise<boolean | undefined>
}

const quote = (arg: string): string => /\s/.test(arg) ? `"${arg}"` : arg

/**
 * Runs a command of smith that has a screen of its own: asks its questions
 * (the options come from smith, none is held here), then either shows its
 * output, or shows the plan (`--dry-run`) or the consequence and asks before it
 * changes anything.
 * @param command - which command, its steps and the arguments the answers build
 * @param deps - how to run smith and ask the person
 * @param preset - answers already known (the app of the row the person pressed a button on); their steps are not asked
 * @returns how it ended; it never throws
 */
export async function runGuidedCommand (command: GuidedCommand, deps: GuidedDeps, preset: GuidedAnswers = {}): Promise<GuidedOutcome> {
  const answers: GuidedAnswers = { ...preset }
  let apps: WebApp[] | undefined
  const webApps = async (): Promise<WebApp[]> => {
    if (apps !== undefined) return apps
    const listed = await deps.run(['environments', 'list', '--json'])
    apps = listed.code === 0 ? parseWebApps(listed.stdout) : []

    return apps
  }

  for (const step of command.steps) {
    if (step.when?.(answers) === false || answers[step.key] !== undefined) continue
    const resolved = await resolve(step, answers, deps, webApps)
    if ('refused' in resolved) return { kind: 'refused', message: resolved.refused }
    const answer = await deps.ask(resolved)
    if (answer === undefined) return { kind: 'cancelled' }
    answers[step.key] = answer
  }

  const args = command.args(answers)
  if (command.mode === 'read') {
    const result = await deps.run(args)
    if (result.code !== 0) return { kind: 'refused', message: describeFailure(result) }
    await deps.show(command.title, result.stdout)

    return { kind: 'shown' }
  }

  const line = ['smith', ...args.map((arg) => quote(arg))].join(' ')
  let detail = `${line}\n\n${command.consequence ?? ''}`
  if (command.planFlag !== undefined) {
    const planned = await deps.run([...args, command.planFlag])
    if (planned.code !== 0) return { kind: 'refused', message: describeFailure(planned) }
    detail = say('vscode.guided-commands.nothing-was-changed-yet', { line, trim: planned.stdout.trim(), part: command.consequence === undefined ? '' : ` ${command.consequence}` })
  }
  // The box comes before the confirmation, unticked: smith commits only when it is ticked.
  const commit = command.commit === true && deps.askCommit !== undefined ? await deps.askCommit() : false
  if ((commit === undefined) || !await deps.confirm(say('vscode.guided-commands.monosmith', { title: command.title }), detail.trim(), command.verb ?? say('vscode.compose-stack.run-it'))) return { kind: 'cancelled' }
  const done = await deps.run([...(command.yes === true ? [...args, '--yes'] : args), ...(commit ? ['--commit'] : [])])

  // A warning smith prints on stderr (the projects a plugin leaves) is part of the answer.
  const lines = `${done.stdout}\n${done.stderr}`.split(/\r?\n/).map((text) => text.trim()).filter(Boolean)

  return done.code === 0 ? { kind: 'done', lines } : { kind: 'refused', message: describeFailure(done) }
}

/** Reads the settings once: `smith config list --json`. */
async function readSettings (deps: GuidedDeps): Promise<Setting[] | string> {
  const listed = await deps.run(['config', 'list', '--json'])

  return listed.code === 0 ? parseSettings(listed.stdout) : describeFailure(listed)
}

/** Reads one project's view once: `smith config list -p <project> --json`. */
async function readProjectView (deps: GuidedDeps, project: string): Promise<ProjectView | string> {
  const listed = await deps.run(['config', 'list', flag.project, project, '--json'])

  return listed.code === 0 ? parseProjectView(listed.stdout) ?? say('vscode.guided-commands.smith-has-no-setting', { setting: project }) : describeFailure(listed)
}

/** The last part of a setting's path: `tag_prefix` of `projects.web.tag_prefix`. */
const lastKey = (path: string): string => path.slice(path.lastIndexOf('.') + 1)

/**
 * Makes the value step fit the setting: true or false, a pick or tick boxes of its fixed values (each explained,
 * the default marked), or a text box with the value now. A project setting that is set can also be cleared.
 * @param resolved - the step to fill
 * @param setting - the setting, as smith listed it
 * @param clearable - true for a project setting: an empty text, or the Clear answer, removes it
 */
function valueStep (resolved: ResolvedStep, setting: Setting, clearable: boolean): void {
  const current = setting.set ? settingWords(setting.value) : []
  const marked = (value: string): string => current.includes(value) ? ' (current)' : ''
  const clear = clearable && setting.set ? [{ label: say('vscode.guided-commands.clear-to-default'), value: '', description: say('vscode.guided-commands.clear-to-default-hint') }] : []
  if (setting.type === 'boolean') {
    resolved.kind = 'pick'
    resolved.options = [...['true', 'false'].map((value) => ({ label: `${value}${marked(value)}`, value, description: setting.reason ?? setting.description })), ...clear]
  } else if (setting.options.length > 0) {
    // Each value says what it does and the default is marked; a list ticks several, starting from what it holds now (or its default).
    const many = setting.type === 'array'
    resolved.kind = many ? 'multi' : 'pick'
    resolved.options = [...setting.options.map((option) => ({
      label:       `${option.value}${option.default ? ` ${say('vscode.guided-commands.default')}` : ''}${many ? '' : marked(option.value)}`,
      value:       option.value,
      description: option.description === '' ? setting.description : option.description,
      picked:      many ? (setting.set ? current.includes(option.value) : option.default) : undefined,
    })), ...clear]
  } else {
    resolved.value = clearable ? current.join(', ') : setting.value
    resolved.optional = clearable
    resolved.prompt = say('vscode.guided-commands.a-list-is-comma-separated', { description: setting.description, part: setting.reason === undefined ? '' : say('vscode.guided-commands.why-the-default-is-what-it', { reason: setting.reason }) })
    if (clearable) resolved.prompt += ` ${say('vscode.guided-commands.empty-clears-it')}`
  }
}

/** Reads the options or the first value of a step from smith. */
async function resolve (step: GuidedStep, answers: GuidedAnswers, deps: GuidedDeps, webApps: () => Promise<WebApp[]>): Promise<ResolvedStep | { refused: string }> {
  const resolved: ResolvedStep = { ...step, options: step.options ?? [] }
  switch (step.from) {
    case 'web-apps': {
      const apps = await webApps()
      if (apps.length === 0) return { refused: say('vscode.guided-commands.no-web-app-here-is-built') }
      resolved.options = apps.map((app) => ({ label: app.project, value: app.project, description: app.environments.map((env) => env.name).join(', ') }))
      break
    }
    case 'environments-of-app': {
      const known = await webApps()
      const app = known.find((item) => item.project === answers.project)
      resolved.options = app === undefined ? [] : environmentsOfApp(app)
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.has-no-environment-that-can-be', { project: String(answers.project) }) }
      break
    }
    case 'workspace-environments': {
      const read = await deps.run(['config', 'get', 'workspace.environments'])
      resolved.value = parseWorkspaceEnvironments(read.code === 0 ? read.stdout : '').join(', ')
      break
    }
    case 'pipeline-templates': {
      const listed = await deps.run(['pipeline'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parsePipelineTemplates(listed.stdout)
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.smith-listed-no-pipeline-update-smith') }
      break
    }
    case 'settings':
    case 'settings-set': {
      const settings = await readSettings(deps)
      if (typeof settings === 'string') return { refused: settings }
      resolved.options = settings
        .filter((setting) => step.from === 'settings' || setting.set)
        .map((setting) => ({ label: setting.path, value: setting.path, description: `${setting.set ? setting.value : say('vscode.guided-commands.default')}: ${setting.description}` }))
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.no-setting-is-written-in-smith') }
      break
    }
    case 'setting-value': {
      const settings = await readSettings(deps)
      if (typeof settings === 'string') return { refused: settings }
      const setting = settings.find((item) => item.path === answers.setting)
      if (setting === undefined) return { refused: say('vscode.guided-commands.smith-has-no-setting', { setting: String(answers.setting) }) }
      valueStep(resolved, setting, false)
      break
    }
    case 'projects': {
      const listed = await deps.run(['projects', '--json'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseProjectNames(listed.stdout)
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.smith-lists-no-project-here') }
      break
    }
    case 'project-settings': {
      const view = await readProjectView(deps, String(answers.project))
      if (typeof view === 'string') return { refused: view }
      // What the manifest owns comes first, read-only and with its file; then the settings of [projects.<name>].
      resolved.options = [
        ...view.facts.map((fact) => ({ label: fact.path, value: `fact:${fact.path}`, description: `${fact.value} · ${say('vscode.guided-commands.read-only-from', { file: fact.source })}` })),
        ...view.settings.map((setting) => ({ label: lastKey(setting.path), value: setting.path, description: `${setting.set ? setting.value : say('vscode.guided-commands.default')}: ${setting.description}` })),
      ]
      break
    }
    case 'project-setting-value': {
      const view = await readProjectView(deps, String(answers.project))
      if (typeof view === 'string') return { refused: view }
      const picked = String(answers.setting)
      const fact = view.facts.find((item) => `fact:${item.path}` === picked)
      if (fact !== undefined) return { refused: say('vscode.guided-commands.comes-from-change-it-there', { key: fact.path, file: fact.source }) }
      const setting = view.settings.find((item) => item.path === picked)
      if (setting === undefined) return { refused: say('vscode.guided-commands.smith-has-no-setting', { setting: picked }) }
      valueStep(resolved, setting, true)
      break
    }
    case 'builtin-plugins': {
      const listed = await deps.run(['plugin', 'list', '--json'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseBuiltinPlugins(listed.stdout)
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.smith-listed-no-built-in-plugin') }
      break
    }
    case 'marketplace': {
      const term = typeof answers.term === 'string' ? answers.term.split(/\s+/).filter(Boolean) : []
      const listed = await deps.run(['plugin', 'search', ...term, '--json'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseMarketplace(listed.stdout)
      if (resolved.options.length === 0) return { refused: say(term.length > 0 ? 'vscode.guided-commands.no-plugin-in-the-marketplace-matches' : 'vscode.guided-commands.the-marketplace-lists-no-plugins-yet') }
      break
    }
    case 'architectures': {
      const listed = await deps.run(['architecture'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseArchitectures(listed.stdout)
      if (resolved.options.length === 0) return { refused: say('vscode.guided-commands.smith-listed-no-architecture') }
      break
    }
    case undefined: {
      break
    }
  }

  return resolved
}
