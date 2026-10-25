import { protocolSchema } from '../protocol/index.ts'
import type { Capabilities, CapabilityName } from '../protocol/index.ts'
import { PLUGIN_MEMBERS } from './plugin-members.config.ts'
import type { SmithPlugin } from './plugin.contract.ts'

/** Whether the plugin implements a method of the protocol: it has every member the method needs, as functions. */
export function implementsMethod (plugin: SmithPlugin, method: string): boolean {
  const entry = Object.hasOwn(PLUGIN_MEMBERS, method) ? PLUGIN_MEMBERS[method] : undefined

  return entry !== undefined && entry.requires.every(member => typeof plugin[member] === 'function')
}

/**
 * The capabilities a plugin advertises, derived from the methods it
 * implements: a plugin cannot advertise a method it lacks, or implement one
 * it does not advertise. The capability of each method is read from the
 * schema's methods table; `cancel` comes from the plugin saying it is
 * `cancellable`.
 */
export function capabilitiesOf (plugin: SmithPlugin): Capabilities {
  const capabilities: Capabilities = {}
  const methods = Object.entries(protocolSchema['x-methods'])
  for (const [method, row] of methods) {
    if (row.capability !== null && implementsMethod(plugin, method)) capabilities[row.capability as CapabilityName] = true
  }
  if (plugin.cancellable === true) capabilities.cancel = true

  return capabilities
}

/** Methods a plugin half-implements: a member of a pair (`install` without `uninstall`) that is not advertised for lack of the other. */
export function halfImplemented (plugin: SmithPlugin): string[] {
  return Object.entries(PLUGIN_MEMBERS)
    .filter(([method, entry]) => typeof plugin[entry.member] === 'function' && !implementsMethod(plugin, method))
    .map(([, entry]) => entry.member)
}
