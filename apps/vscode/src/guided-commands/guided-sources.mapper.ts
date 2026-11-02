import type { GuidedOption } from './guided-command.contract.ts'

/** A web app built once per environment, as `smith environments list --json` prints it. */
export interface WebApp {
  project:      string
  default:      string
  environments: Array<{ name: string, task: string }>
}

/** The environments of a new app when the workspace names none: lcl is the developer's machine. */
const DEFAULT_ENVIRONMENTS = ['lcl', 'dev', 'uat', 'prd']

/**
 * Reads `smith environments list --json`.
 * @param json - what smith printed
 * @returns the web apps; an unreadable text gives none
 */
export function parseWebApps (json: string): WebApp[] {
  let value: unknown
  try {
    value = JSON.parse(json)
  } catch {
    return []
  }
  if (!Array.isArray(value)) return []

  const records = value.filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)

  return records.filter((record) => typeof record.project === 'string').map((record) => {
    const environments = (Array.isArray(record.environments) ? record.environments : []) as Array<Record<string, unknown>>

    return {
      project:      record.project as string,
      default:      typeof record.default === 'string' ? record.default : '',
      environments: environments
        .filter((env) => typeof env.name === 'string')
        .map((env) => ({ name: env.name as string, task: typeof env.task === 'string' ? env.task : '' })),
    }
  })
}

/**
 * The environments of an app that `environments remove` accepts, as answers: the default one cannot be removed.
 * @param app - the web app
 */
export function environmentsOfApp (app: WebApp): GuidedOption[] {
  return app.environments
    .filter((env) => env.name !== app.default)
    .map((env) => ({ label: env.name, value: env.name, description: `built by smith run ${env.task}; its .env file stays` }))
}

/**
 * The names `smith config get workspace.environments` printed, or the default four when none is set.
 * @param text - what smith printed (empty when the setting is not set)
 */
export function parseWorkspaceEnvironments (text: string): string[] {
  const names = text.match(/[a-z][a-z0-9-]*/g) ?? []

  return names.length > 0 ? names : DEFAULT_ENVIRONMENTS
}

/**
 * The pipelines `smith pipeline` lists: the name, and what it does with the app it needs when it needs one.
 * @param text - the table smith printed: name, providers, description, in columns split by two spaces
 */
export function parsePipelineTemplates (text: string): GuidedOption[] {
  const options: GuidedOption[] = []
  for (const line of text.split(/\r?\n/)) {
    const [first = '', providers = '', ...rest] = line.trim().split(/\s{2,}/)
    const name = /^[a-z][a-z0-9-]*/.exec(first)?.[0]
    if (name === undefined) continue
    const needsApp = first.includes('--project') ? ' (asks for the app)' : ''
    options.push({ label: name, value: name, description: `${rest.join(' ')}${needsApp}${providers === '' ? '' : ` [${providers}]`}`.trim() })
  }

  return options
}
