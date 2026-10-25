import type { RpcError } from '../protocol/index.ts'

/** Why the host stopped waiting for a call. */
export type InterruptionReason = 'cancelled' | 'timeout'

/**
 * A call the host gave up on, because its `signal` aborted (`cancelled`) or its
 * time ran out (`timeout`). `killed` says the plugin did not stop in time
 * and was killed, which also fails every other call in flight to it. When the
 * plugin answered while being stopped, `response` is its error (the
 * request-cancelled code, usually).
 */
export class CallInterruptedError extends Error {
  readonly reason:    InterruptionReason
  readonly killed:    boolean
  readonly response?: RpcError

  constructor (message: string, reason: InterruptionReason, killed: boolean, response?: RpcError) {
    super(message, response === undefined ? undefined : { cause: response })
    this.name = 'CallInterruptedError'
    this.reason = reason
    this.killed = killed
    if (response !== undefined) this.response = response
  }
}
