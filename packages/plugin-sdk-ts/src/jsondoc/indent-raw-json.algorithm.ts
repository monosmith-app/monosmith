/** A string literal (escapes included), a bracket, a comma or colon, or a run of any other characters (a number, `true`, `null`). */
const TOKEN = /"(?:[^"\\]|\\.)*"|[{}[\],:]|[^\s{}[\],:"]+/g

/**
 * Formats JSON text as Go's `json.Indent(dst, src, prefix, "  ")` does: the
 * insignificant whitespace is dropped and every element goes on its own line,
 * indented two spaces per level and starting with `prefix`; an empty array or
 * object stays `[]` or `{}`. Text that is not an array is returned as it is. The text
 * of leaves (numbers, strings) is kept exactly.
 */
export function indentRawJson (raw: string, prefix: string): string {
  if (!raw.startsWith('[')) return raw
  let out = ''
  let depth = 0
  // Right after an opening bracket: the line break is held back until it is known whether the container is empty.
  let needIndent = false
  const newline = (): void => {
    out += `\n${prefix}${'  '.repeat(depth)}`
  }
  const open = (token: string): void => {
    needIndent = true
    out += token
  }
  const close = (token: string): void => {
    if (needIndent) {
      needIndent = false
    } else {
      depth--
      newline()
    }
    out += token
  }
  const handlers: Readonly<Record<string, (token: string) => void>> = {
    '{': open,
    '[': open,
    '}': close,
    ']': close,
    ',': (token) => {
      out += token
      newline()
    },
    ':': () => {
      out += ': '
    },
  }

  for (const [token] of raw.matchAll(TOKEN)) {
    if (needIndent && token !== '}' && token !== ']') {
      needIndent = false
      depth++
      newline()
    }
    const handler = Object.hasOwn(handlers, token) ? handlers[token] : undefined
    if (handler === undefined) out += token
    else handler(token)
  }

  return out
}
