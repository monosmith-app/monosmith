import { PROTOCOL_VERSION } from '../protocol/index.ts'
import type { InitializeParams, InitializeResult } from '../protocol/index.ts'
import { capabilitiesOf } from './capabilities-of.algorithm.ts'
import type { SmithPlugin } from './plugin.contract.ts'

/**
 * The handshake. It refuses a host that does not speak this SDK's protocol
 * version (the message is the Go SDK's), asks the plugin to describe itself,
 * and fills in the version it chose and the capabilities derived from the
 * methods the plugin has, whatever `describe` returned for them.
 */
export async function initializePlugin (plugin: SmithPlugin, params: InitializeParams): Promise<InitializeResult> {
  if (!params.protocolVersions.includes(PROTOCOL_VERSION)) {
    // The Go SDK formats the host's list with %v: [1 2].
    throw new Error(`no common protocol version: host speaks [${params.protocolVersions.join(' ')}], plugin speaks ${PROTOCOL_VERSION}`)
  }
  const description = await plugin.describe(params)

  return { ...description, protocolVersion: PROTOCOL_VERSION, capabilities: capabilitiesOf(plugin) }
}
