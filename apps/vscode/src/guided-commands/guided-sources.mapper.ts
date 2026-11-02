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

/** A setting of smith.toml, as `smith config list --json` prints it. */
export interface Setting {
  path:        string
  type:        string
  description: string
  enum?:       string[]
  reason?:     string
  value:       string
  set:         boolean
}

/**
 * Reads `smith config list --json`.
 * @param json - what smith printed
 * @returns the settings that have a path; an unreadable text gives none
 */
export function parseSettings (json: string): Setting[] {
  let value: unknown
  try {
    value = JSON.parse(json)
  } catch {
    return []
  }
  if (!Array.isArray(value)) return []
  const records = value.filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)

  return records.filter((record) => typeof record.path === 'string').map((record) => ({
    path:        record.path as string,
    type:        typeof record.type === 'string' ? record.type : '',
    description: typeof record.description === 'string' ? record.description : '',
    enum:        Array.isArray(record.enum) ? record.enum.filter((item): item is string => typeof item === 'string') : undefined,
    reason:      typeof record.reason === 'string' && record.reason !== '' ? record.reason : undefined,
    value:       typeof record.value === 'string' ? record.value : '',
    set:         record.set === true,
  }))
}

/**
 * The built-in plugins `smith plugin list --json` prints.
 * @param json - what smith printed
 */
export function parseBuiltinPlugins (json: string): GuidedOption[] {
  let value: unknown
  try {
    value = JSON.parse(json)
  } catch {
    return []
  }
  if (!Array.isArray(value)) return []
  const records = value.filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)

  return records
    .filter((record) => typeof record.name === 'string' && record.source === 'built-in')
    .map((record) => ({ label: record.name as string, value: record.name as string, description: Array.isArray(record.provides) ? record.provides.join(', ') : undefined }))
}

/**
 * The plugins `smith plugin search --json` found: the newest version, what it does and its roles.
 * @param json - what smith printed (a sentence when the marketplace lists nothing)
 */
export function parseMarketplace (json: string): GuidedOption[] {
  let value: unknown
  try {
    value = JSON.parse(json)
  } catch {
    return []
  }
  if (!Array.isArray(value)) return []
  const records = value.filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)

  return records.filter((record) => typeof record.name === 'string').map((record) => {
    const roles = Array.isArray(record.roles) && record.roles.length > 0 ? ` [${record.roles.join(', ')}]` : ''
    const newest = typeof record.newest === 'string' && record.newest !== '' ? `v${record.newest} · ` : ''

    return { label: record.name as string, value: record.name as string, description: `${newest}${typeof record.description === 'string' ? record.description : ''}${roles}` }
  })
}

/**
 * The architectures `smith architecture` lists: a name and what it means, the one in use marked.
 * @param text - what smith printed
 */
export function parseArchitectures (text: string): GuidedOption[] {
  const options: GuidedOption[] = []
  for (const line of text.split(/\r?\n/)) {
    const match = /^([ *]) ([a-z][a-z0-9-]*)(?= )/.exec(line)
    if (match === null) continue
    options.push({ label: `${match[2]}${match[1] === '*' ? ' (in use)' : ''}`, value: match[2], description: line.slice(match[0].length).trim() })
  }

  return options
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
