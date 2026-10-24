/**
 * Waits until a condition holds, polling it, and fails with what was waited
 * for when it never does.
 * @param what - what is waited for, for the failure message
 * @param condition - true when the wait is over
 * @param timeoutMs - how long to wait at most
 */
export async function waitUntil (what: string, condition: () => boolean, timeoutMs = 30_000): Promise<void> {
  const deadline = Date.now() + timeoutMs
  while (!condition()) {
    if (Date.now() > deadline) throw new Error(`timed out waiting for ${what}`)
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
}
