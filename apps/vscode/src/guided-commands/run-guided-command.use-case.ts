import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { GuidedAnswers, GuidedCommand, GuidedOutcome, GuidedStep, ResolvedStep } from './guided-command.contract.ts'
import { environmentsOfApp, parseArchitectures, parseBuiltinPlugins, parseMarketplace, parsePipelineTemplates, parseSettings, parseWebApps, parseWorkspaceEnvironments, type Setting, type WebApp } from './guided-sources.mapper.ts'

/** What running a guided command needs from the editor. */
export interface GuidedDeps {
  /** Runs smith in the workspace folder. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Asks one step. Resolves the answer (a list for `multi`), or undefined when the person cancels. */
  ask:     (step: ResolvedStep) => Promise<string | string[] | undefined>
  /** Shows the text of a `read` command. */
  show:    (title: string, text: string) => Promise<void>
  /** Shows what will change and asks. Nothing is changed before a yes. */
  confirm: (title: string, detail: string, verb: string) => Promise<boolean>
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
    detail = `${line}\n\n${planned.stdout.trim()}\n\nNothing was changed yet.${command.consequence === undefined ? '' : ` ${command.consequence}`}`
  }
  if (!await deps.confirm(`MonoSmith: ${command.title}?`, detail.trim(), command.verb ?? 'Run it')) return { kind: 'cancelled' }
  const done = await deps.run(command.yes === true ? [...args, '--yes'] : args)

  // A warning smith prints on stderr (the projects a plugin leaves) is part of the answer.
  const lines = `${done.stdout}\n${done.stderr}`.split(/\r?\n/).map((text) => text.trim()).filter(Boolean)

  return done.code === 0 ? { kind: 'done', lines } : { kind: 'refused', message: describeFailure(done) }
}

/** Reads the settings once: `smith config list --json`. */
async function readSettings (deps: GuidedDeps): Promise<Setting[] | string> {
  const listed = await deps.run(['config', 'list', '--json'])

  return listed.code === 0 ? parseSettings(listed.stdout) : describeFailure(listed)
}

/** Reads the options or the first value of a step from smith. */
async function resolve (step: GuidedStep, answers: GuidedAnswers, deps: GuidedDeps, webApps: () => Promise<WebApp[]>): Promise<ResolvedStep | { refused: string }> {
  const resolved: ResolvedStep = { ...step, options: step.options ?? [] }
  switch (step.from) {
    case 'web-apps': {
      const apps = await webApps()
      if (apps.length === 0) return { refused: 'No web app here is built per environment. smith add app <name> --with react --environments lcl,dev,uat,prd makes one.' }
      resolved.options = apps.map((app) => ({ label: app.project, value: app.project, description: app.environments.map((env) => env.name).join(', ') }))
      break
    }
    case 'environments-of-app': {
      const known = await webApps()
      const app = known.find((item) => item.project === answers.project)
      resolved.options = app === undefined ? [] : environmentsOfApp(app)
      if (resolved.options.length === 0) return { refused: `${String(answers.project)} has no environment that can be removed: the default one stays.` }
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
      if (resolved.options.length === 0) return { refused: 'smith listed no pipeline. Update smith, or run smith pipeline in a terminal.' }
      break
    }
    case 'settings':
    case 'settings-set': {
      const settings = await readSettings(deps)
      if (typeof settings === 'string') return { refused: settings }
      resolved.options = settings
        .filter((setting) => step.from === 'settings' || setting.set)
        .map((setting) => ({ label: setting.path, value: setting.path, description: `${setting.set ? setting.value : '(default)'}: ${setting.description}` }))
      if (resolved.options.length === 0) return { refused: 'No setting is written in smith.toml, so there is nothing to reset.' }
      break
    }
    case 'setting-value': {
      const settings = await readSettings(deps)
      if (typeof settings === 'string') return { refused: settings }
      const setting = settings.find((item) => item.path === answers.setting)
      if (setting === undefined) return { refused: `smith has no setting ${String(answers.setting)}.` }
      const marked = (value: string): string => value === setting.value && setting.set ? ' (current)' : ''
      const choices = setting.type === 'boolean' ? ['true', 'false'] : (setting.enum ?? [])
      if (choices.length > 0) {
        resolved.kind = 'pick'
        resolved.options = choices.map((value) => ({ label: `${value}${marked(value)}`, value, description: setting.reason ?? setting.description }))
      } else {
        resolved.value = setting.value
        resolved.prompt = `${setting.description}${setting.reason === undefined ? '' : ` Why the default is what it is: ${setting.reason}`} A list is comma-separated.`
      }
      break
    }
    case 'builtin-plugins': {
      const listed = await deps.run(['plugin', 'list', '--json'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseBuiltinPlugins(listed.stdout)
      if (resolved.options.length === 0) return { refused: 'smith listed no built-in plugin.' }
      break
    }
    case 'marketplace': {
      const term = typeof answers.term === 'string' ? answers.term.split(/\s+/).filter(Boolean) : []
      const listed = await deps.run(['plugin', 'search', ...term, '--json'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseMarketplace(listed.stdout)
      if (resolved.options.length === 0) return { refused: term.length > 0 ? 'No plugin in the marketplace matches. Search with fewer words, or leave the field empty to list them all.' : 'The marketplace lists no plugins yet, so there is nothing to install.' }
      break
    }
    case 'architectures': {
      const listed = await deps.run(['architecture'])
      if (listed.code !== 0) return { refused: describeFailure(listed) }
      resolved.options = parseArchitectures(listed.stdout)
      if (resolved.options.length === 0) return { refused: 'smith listed no architecture.' }
      break
    }
    case undefined: {
      break
    }
  }

  return resolved
}
