import { EncodeError, ParamsError, decodeParams, encodeResult } from '../codec/index.ts'
import type { ReceivedRequest } from '../codec/index.ts'
import { ERROR_CODES } from '../protocol/index.ts'
import type { InitializeParams } from '../protocol/index.ts'
import { implementsMethod } from './capabilities-of.algorithm.ts'
import { initializePlugin } from './initialize-plugin.use-case.ts'
import { PLUGIN_MEMBERS } from './plugin-members.config.ts'
import type { SmithPlugin } from './plugin.contract.ts'
import type { RequestOutcome } from './request-outcome.contract.ts'

/** The text of an exception: its message, or what was thrown. */
export function messageOf (thrown: unknown): string {
  return thrown instanceof Error ? thrown.message : String(thrown)
}

/**
 * Answers one request: `initialize` and `shutdown` here, every other method
 * by the member of the plugin that implements it. The codes are the Go SDK's:
 * MethodNotFound for a method the plugin does not implement, InvalidParams
 * for params of the wrong shape, PluginFailure for an exception the method
 * threw (its message is shown to the user), InternalError for a result that
 * cannot be written as the schema describes. `signal` is passed to the method as its
 * second argument; it aborts when the host cancels the request.
 */
export async function dispatchRequest (plugin: SmithPlugin, request: ReceivedRequest, signal: AbortSignal = new AbortController().signal): Promise<RequestOutcome> {
  if (request.method === 'shutdown') return { ok: true, resultJson: '{}' }

  let params: unknown
  if (request.method === 'initialize' || implementsMethod(plugin, request.method)) {
    try {
      params = decodeParams(request.method, request.params)
    } catch (error) {
      if (error instanceof ParamsError) return { ok: false, code: ERROR_CODES.InvalidParams, message: error.message }
      throw error
    }
  }

  let result: unknown
  try {
    if (request.method === 'initialize') {
      result = await initializePlugin(plugin, params as InitializeParams)
    } else if (implementsMethod(plugin, request.method)) {
      const member = PLUGIN_MEMBERS[request.method].member
      result = await (plugin[member] as (params: unknown, signal: AbortSignal) => unknown).call(plugin, params, signal)
    } else {
      return { ok: false, code: ERROR_CODES.MethodNotFound, message: `method not supported: ${request.method}` }
    }
  } catch (error) {
    return { ok: false, code: ERROR_CODES.PluginFailure, message: messageOf(error) }
  }

  try {
    return { ok: true, resultJson: encodeResult(request.method, result) }
  } catch (error) {
    if (error instanceof EncodeError) return { ok: false, code: ERROR_CODES.InternalError, message: `${request.method}: ${error.message}` }
    throw error
  }
}
