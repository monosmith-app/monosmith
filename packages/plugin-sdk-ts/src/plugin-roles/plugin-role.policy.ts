import { protocolSchema } from '../protocol/index.ts'
import type { InitializeResult, PluginKind, Role } from '../protocol/index.ts'

/** Every known role, in the order a project's make-up lists them. */
export const ROLES: readonly Role[] = protocolSchema['x-roles'].map(row => row.name as Role)

/** Whether the role is one of {@link ROLES}. */
export function isKnownRole (role: string): role is Role {
  return (ROLES as readonly string[]).includes(role)
}

/** Whether a project has at most one plugin at this level (language, scaffold). A toolchain is exclusive per slot instead. */
export function isExclusiveRole (role: Role): boolean {
  return cardinalityOf(role) === 'exclusive'
}

/** Whether any number of plugins can share the level, each only adding (structure, deploy). */
export function isAdditiveRole (role: Role): boolean {
  return cardinalityOf(role) === 'additive'
}

function cardinalityOf (role: Role): string | undefined {
  return protocolSchema['x-roles'].find(row => row.name === role)?.cardinality
}

/**
 * The roles of a plugin that declares none: a language plugin brings its own
 * default toolchain and scaffold; any other kind of plugin is a structure rule.
 */
export function defaultRoles (kind: PluginKind): Role[] {
  const wanted = kind === 'language' ? 'language' : 'convention'

  return protocolSchema['x-roles'].filter(row => row.defaultFor.includes(wanted)).map(row => row.name as Role)
}

/** The plugin's roles, or the default of its kind. */
export function effectiveRoles (info: Pick<InitializeResult, 'kind' | 'roles'>): Role[] {
  return info.roles !== undefined && info.roles.length > 0 ? info.roles : defaultRoles(info.kind)
}
