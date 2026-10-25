/**
 * A JSON-RPC 2.0 error object as an exception: what a host's call rejects
 * with when the plugin answers an error, and what the serve loop turns an
 * exception back into. `code` is one of {@link ERROR_CODES}.
 */
export class RpcError extends Error {
  readonly code: number

  constructor (code: number, message: string) {
    super(message)
    this.name = 'RpcError'
    this.code = code
  }
}
