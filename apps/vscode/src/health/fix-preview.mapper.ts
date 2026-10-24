/** What `smith doctor --fix` proposes, before anything is written. */
export interface FixPreview {
  /** One line per fix, such as `ignore the cache directory`. */
  titles: string[]
  /** The fixes with their diffs, as smith prints them. */
  text:   string
}

const MARKER = /^● (.+)$/

/**
 * Reads the proposed fixes out of what `doctor.fixes` answers. smith prints
 * them for a person and the method carries that text in the message of its
 * answer (the command exits without JSON), so the fixes are the lines that
 * start with a bullet, with the diff that follows, up to the line that asks
 * for `--yes`.
 * @param message - the text `doctor.fixes` failed with
 * @returns the fixes, or undefined when smith proposes none
 */
export function parseFixPreview (message: string): FixPreview | undefined {
  const lines = message.split(/\r?\n/)
  const first = lines.findIndex((line) => MARKER.test(line))
  if (first === -1) return undefined
  const end = lines.findIndex((line, index) => index > first && line.startsWith('Pass --yes'))
  const block = lines.slice(first, end === -1 ? undefined : end)
  const titles = block.flatMap((line) => MARKER.exec(line)?.[1] ?? [])

  return { titles, text: block.join('\n').trim() }
}
