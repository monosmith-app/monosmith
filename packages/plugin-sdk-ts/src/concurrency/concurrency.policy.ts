/** The `concurrency` value that leaves the number of simultaneous requests to the host. */
export const CONCURRENCY_UNLIMITED = -1

/** How many requests the host keeps in flight to a plugin that declared unlimited concurrency. */
export const UNLIMITED_IN_FLIGHT = 32

/** The most workers a plugin that declared unlimited concurrency runs: the host decides, but the pool stays bounded. */
export const UNLIMITED_WORKERS = 1024

/** Whether a declared concurrency is one the protocol defines: absent (0), a count (1 and above) or unlimited (-1). */
export function isValidConcurrency (declared: number): boolean {
  return Number.isSafeInteger(declared) && declared >= CONCURRENCY_UNLIMITED
}

/**
 * Reads a declared concurrency: how many requests may be in flight at once,
 * or `unlimited`. An absent, zero or invalid value is 1, the default of every
 * plugin that says nothing.
 */
export function concurrencyLimit (declared: number | undefined): { limit: number, unlimited: boolean } {
  if (declared === CONCURRENCY_UNLIMITED) return { limit: 0, unlimited: true }

  return { limit: declared === undefined || !Number.isSafeInteger(declared) || declared < 1 ? 1 : declared, unlimited: false }
}

/** How many requests the host keeps in flight to a plugin that declared `declared`. */
export function inFlightLimit (declared: number | undefined): number {
  const { limit, unlimited } = concurrencyLimit(declared)

  return unlimited ? UNLIMITED_IN_FLIGHT : limit
}

/** How many requests a plugin that declared `declared` works on at once (its pool size). */
export function workerLimit (declared: number | undefined): number {
  const { limit, unlimited } = concurrencyLimit(declared)

  return unlimited ? UNLIMITED_WORKERS : limit
}
