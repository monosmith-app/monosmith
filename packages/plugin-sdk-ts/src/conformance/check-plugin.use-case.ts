import type { PluginClient } from '../host/index.ts'
import { describeRef, isEmptyRef } from '../plugin-refs/index.ts'
import { ROLES, effectiveRoles, isKnownRole } from '../plugin-roles/index.ts'
import type { PluginRef, Role } from '../protocol/index.ts'
import { checkGeneration } from './check-generation.use-case.ts'
import { checkScheduling } from './check-scheduling.use-case.ts'
import { ConformanceReport } from './conformance-report.model.ts'
import { goList, goQuote } from './go-list.mapper.ts'

/** Options of {@link checkPlugin}. */
export interface CheckOptions {
  /**
   * An empty directory the plugin was started in (its `workspaceRoot`). When
   * set, a plugin that generates projects is asked to generate an app, a lib
   * and a package there, which a language plugin must then discover as
   * generated. Generation may install dependencies, so it is opt-in.
   */
  workspace?: string
}

const CALL_TIMEOUT_MS = 20 * 60 * 1000

/**
 * Runs the checks of `smith plugin check` against a connected plugin: it names
 * itself, has a known kind and known roles, its roles match what it
 * implements, it names the slots it fills, its references and stacks are
 * well-formed, every advertised method answers a request with nothing in it,
 * requests sent together are answered as the declared `concurrency` allows, a
 * plugin that advertises `cancel` stops a cancelled request, and, with `workspace`, the projects it generates are found again by its own
 * `discover`. A port of the Go SDK's `conformance.Check`, with the same checks
 * under the same names. (`smith plugin check` adds one more, "shuts down when
 * asked", after closing the plugin: call {@link PluginClient.close} and add it
 * to the report if you want it.)
 */
export async function checkPlugin (client: PluginClient, options: CheckOptions = {}): Promise<ConformanceReport> {
  const info = client.info
  const report = new ConformanceReport(info.plugin)
  report.add('names itself', info.plugin.name !== '' && info.plugin.version !== '', `initialize returned name ${goQuote(info.plugin.name)}, version ${goQuote(info.plugin.version)}; both are required`)
  report.add('has a known kind', info.kind === 'language' || info.kind === 'convention', `kind ${goQuote(info.kind)} is neither language nor convention`)
  const roles = effectiveRoles(info)
  for (const role of roles) report.add(`plays a known role: ${role}`, isKnownRole(role), `roles are ${goList(ROLES)}`)
  const caps = info.capabilities
  if (roles.includes('language')) {
    report.add('a language plugin discovers projects', caps.discover === true && (info.manifests ?? []).length > 0, 'it must implement discover and name its manifest files')
  }
  if (roles.includes('scaffold')) report.add('a scaffold generates projects', caps.generate === true, 'it must implement generate')
  for (const role of ['toolchain', 'structure', 'deploy'] as const satisfies readonly Role[]) {
    if (roles.includes(role) && !roles.includes('language')) {
      report.add(`a ${role} plugin does something`, caps.contribute === true || caps.generate === true || caps.ci === true || caps.doctor === true, 'it implements none of contribute, generate, ci and doctor')
    }
  }
  const provides = info.provides ?? []
  if (provides.length > 0) report.add('only a toolchain provides tasks', roles.includes('toolchain'), `it provides ${goList(provides)} without the toolchain role`)
  if (info.kind === 'convention' && roles.includes('toolchain')) report.add('a toolchain names the slots it fills', provides.length > 0, 'a toolchain plugin must list the tasks it provides')
  const requires = info.requires ?? []
  const conflicts = info.conflicts ?? []
  const stacks = info.stacks ?? []
  for (const ref of requires) report.add(`requires ${describeRef(ref)}: a known reference`, isValidRef(ref), 'a reference sets a known level, a name or a slot')
  for (const ref of conflicts) {
    report.add(`conflicts with ${describeRef(ref)}: a known reference`, isValidRef(ref) && ref.name !== info.plugin.name, 'a reference sets a known level, a name or a slot, and is not the plugin itself')
  }
  const names = new Set<string>()
  for (const stack of stacks) {
    report.add(`stack ${stack.name} is complete`, stack.name !== '' && stack.language !== '' && !names.has(stack.name), 'a stack needs a unique name and a language plugin')
    names.add(stack.name)
  }

  await checkScheduling(client, report)

  const answers = async (check: string, method: 'discover' | 'contribute' | 'ci' | 'doctor', params: Record<string, unknown>): Promise<void> => {
    if (!client.supports(method)) return
    try {
      await client.call(method, params as never, { timeoutMs: CALL_TIMEOUT_MS })
      report.add(check, true, '')
    } catch (error) {
      report.add(check, false, error instanceof Error ? error.message : String(error))
    }
  }
  await answers('discover answers with no manifests', 'discover', { manifests: [] })
  await answers('contribute answers with no projects', 'contribute', { projects: [] })
  await answers('ci answers with no projects', 'ci', { projects: [] })
  await answers('doctor answers with no projects', 'doctor', { projects: [] })

  if (options.workspace !== undefined && caps.generate === true) await checkGeneration(client, options.workspace, report)

  return report
}

/** Whether a reference names something: a known level, a name or a slot. */
function isValidRef (ref: PluginRef): boolean {
  const level: string = ref.level ?? ''

  return !isEmptyRef(ref) && (level === '' || isKnownRole(level))
}
