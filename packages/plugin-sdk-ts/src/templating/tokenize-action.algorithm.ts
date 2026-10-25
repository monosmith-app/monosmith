export type Token =
  | { t: 'open' | 'close' | 'pipe' } |
  { t: 'field', root: boolean, path: string[] } |
  { t: 'string', value: string } |
  { t: 'number', value: number } |
  { t: 'word', value: string }

/**
 * The elements of an action, in the order the alternatives are tried: space,
 * brackets and pipe, a quoted or raw string, a number, a field (`.A.B`, `$`,
 * `$.A`, or a lone `.`), a word, and `:=`/`=`, which are variables. Anything else
 * is the last alternative, so every character belongs to some match.
 */
const ELEMENT = new RegExp([
  String.raw`\s+`,
  String.raw`(?<open>\()`,
  String.raw`(?<close>\))`,
  String.raw`(?<pipe>\|)`,
  String.raw`(?<string>"(?:[^"\\]|\\.)*")`,
  '(?<raw>`[^`]*`)',
  String.raw`(?<number>-?\d+(?:\.\d+)?(?![\w.]))`,
  String.raw`(?<field>\$(?:\.[A-Z_]\w*)*|(?:\.[A-Z_]\w*)+|\.)`,
  String.raw`(?<word>[A-Z_]\w*)`,
  '(?<variable>:?=)',
  '(?<other>.)',
].join('|'), 'gis')

/**
 * Splits the text of an action into tokens. `fail` reports what cannot be a token.
 */
export function tokenizeAction (body: string, fail: (message: string) => never): Token[] {
  const tokens: Token[] = []
  for (const match of body.matchAll(ELEMENT)) {
    const found = match.groups ?? {}
    if (found.open !== undefined) tokens.push({ t: 'open' })
    else if (found.close !== undefined) tokens.push({ t: 'close' })
    else if (found.pipe !== undefined) tokens.push({ t: 'pipe' })
    else if (found.string !== undefined) tokens.push({ t: 'string', value: unquote(found.string, fail) })
    else if (found.raw !== undefined) tokens.push({ t: 'string', value: found.raw.slice(1, -1) })
    else if (found.number !== undefined) tokens.push({ t: 'number', value: Number(found.number) })
    else if (found.field !== undefined) tokens.push({ t: 'field', root: found.field.startsWith('$'), path: found.field.split('.').filter(part => part !== '' && part !== '$') })
    else if (found.word !== undefined) tokens.push({ t: 'word', value: found.word })
    else if (found.variable !== undefined) fail('variables are not supported by the SDK\'s templates')
    else if (found.other !== undefined) fail(`unexpected character "${found.other}" in action`)
  }

  return tokens
}

function unquote (literal: string, fail: (message: string) => never): string {
  try {
    return JSON.parse(literal) as string
  } catch {
    return fail(`invalid string ${literal}`)
  }
}
