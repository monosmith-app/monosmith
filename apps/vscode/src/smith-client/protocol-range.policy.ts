/** The protocol versions of `smith serve` this extension speaks. */
export const SUPPORTED_PROTOCOLS = { oldest: 1, newest: 1 } as const

/**
 * Whether a smith's protocol is inside the range. Older cannot be served;
 * newer is worked with, since the protocol only ever adds optional fields.
 */
export function judgeProtocol (version: number): 'supported' | 'older' | 'newer' {
  if (version < SUPPORTED_PROTOCOLS.oldest) return 'older'

  return version > SUPPORTED_PROTOCOLS.newest ? 'newer' : 'supported'
}
