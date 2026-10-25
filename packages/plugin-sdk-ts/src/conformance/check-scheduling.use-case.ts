import { CallInterruptedError } from '../host/index.ts'
import type { PluginClient } from '../host/index.ts'
import { ERROR_CODES } from '../protocol/index.ts'
import { inFlightLimit, isValidConcurrency } from '../concurrency/index.ts'
import type { ConformanceReport } from './conformance-report.model.ts'

/** How long a plugin that advertised cancel has to answer a request the host cancelled. */
export const CANCEL_BOUND_MS = 10_000

/** An id no request ever had. */
const UNKNOWN_REQUEST = 2 ** 40

type Probe = (signal?: AbortSignal) => Promise<unknown>

/** The cheapest call the plugin answers without doing work: a method that takes a list of projects or manifests, asked about none. */
function probeOf (client: PluginClient): Probe | undefined {
  const timeoutMs = 2 * 60 * 1000
  if (client.supports('discover')) return signal => client.call('discover', { manifests: [] }, { timeoutMs, signal })
  if (client.supports('contribute')) return signal => client.call('contribute', { projects: [] }, { timeoutMs, signal })
  if (client.supports('ci')) return signal => client.call('ci', { projects: [] }, { timeoutMs, signal })

  return client.supports('doctor') ? signal => client.call('doctor', { projects: [] }, { timeoutMs, signal }) : undefined
}

function messageOf (thrown: unknown): string {
  return thrown instanceof Error ? thrown.message : String(thrown)
}

/**
 * Checks what a plugin says about how it is called: the concurrency it
 * declares must be one the protocol defines and be accepted, and the cancel it
 * advertises must stop a request within a bound and leave the plugin serving.
 * The same checks, under the same names, as the Go SDK's.
 */
export async function checkScheduling (client: PluginClient, report: ConformanceReport, cancelBoundMs = CANCEL_BOUND_MS): Promise<void> {
  const info = client.info
  report.add('declares a concurrency the protocol defines', isValidConcurrency(info.concurrency ?? 0), `concurrency ${info.concurrency} is below -1; it must be absent, 1 or more, or -1 for unlimited`)
  const probe = probeOf(client)
  if (probe === undefined) return
  await checkConcurrency(client, probe, report)
  if (info.capabilities.cancel === true) await checkCancel(client, probe, report, cancelBoundMs)
}

/**
 * Sends requests together. A plugin that declares nothing is sent them one at
 * a time and must answer them all; one that declares a concurrency is sent that
 * many at once (at most four, to keep the check cheap) and must answer them all.
 */
async function checkConcurrency (client: PluginClient, probe: Probe, report: ConformanceReport): Promise<void> {
  const limit = inFlightLimit(client.info.concurrency)
  let together = Math.min(limit, 4)
  let name = `accepts ${together} requests at the same time`
  if (limit === 1) {
    together = 3
    name = 'answers 3 requests sent together, one at a time'
  }
  const settled = await Promise.allSettled(Array.from({ length: together }, async () => await probe()))
  const failed = settled.find((result): result is PromiseRejectedResult => result.status === 'rejected')
  const peak = client.peakInFlight
  report.add(name, failed === undefined && peak <= limit, `${failed === undefined ? '<nil>' : messageOf(failed.reason)} (the host had ${peak} in flight, the plugin declared ${limit})`)
}

/** How a promise ended: with the error it rejected with, if any. */
async function settle (promise: Promise<unknown>): Promise<{ error?: unknown }> {
  try {
    await promise

    return {}
  } catch (error) {
    return { error }
  }
}

/** Cancels a request, then a request that does not exist. */
async function checkCancel (client: PluginClient, probe: Probe, report: ConformanceReport, cancelBoundMs: number): Promise<void> {
  const controller = new AbortController()
  const began = Date.now()
  const request = settle(probe(controller.signal))
  await new Promise(resolve => setTimeout(resolve, 5)) // let the request leave; a cancel that wins the race is just as good an answer
  controller.abort()
  let timer: NodeJS.Timeout | undefined
  const late = new Promise<'late'>((resolve) => {
    timer = setTimeout(() => { resolve('late') }, cancelBoundMs + client.cancelGraceMs)
  })
  const outcome = await Promise.race([request, late])
  clearTimeout(timer)
  const error = outcome === 'late' ? new Error(`no answer within ${cancelBoundMs}ms of the cancel`) : outcome.error
  const stopped = outcome !== 'late' && (error === undefined || (error instanceof CallInterruptedError && !error.killed))
  const detail = `${error === undefined ? '<nil>' : messageOf(error)} (a plugin that advertised cancel answers the request, with ${ERROR_CODES.RequestCancelled} if it stopped, within ${cancelBoundMs}ms)`
  report.add('answers a cancelled request with its result or the cancelled error', stopped && Date.now() - began < cancelBoundMs, detail)

  let notifyError: unknown
  let followUp: unknown
  try {
    await client.notify('cancel', { id: UNKNOWN_REQUEST })
  } catch (error_) {
    notifyError = error_
  }
  try {
    await probe()
  } catch (error_) {
    followUp = error_
  }
  report.add('ignores a cancel for a request it does not know', notifyError === undefined && followUp === undefined,
    `${notifyError === undefined ? '<nil>' : messageOf(notifyError)}, then ${followUp === undefined ? '<nil>' : messageOf(followUp)}`)
}
