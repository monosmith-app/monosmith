export type TemplateFunction = (args: unknown[]) => unknown

export function isObject (value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

/** Keys in the order Go sorts the keys of a map in. */
export function byKey (a: string, b: string): number {
  return Number(a > b) - Number(a < b)
}

/** Go's truth: false, 0, nil, and an empty string, array or map are false; everything else is true. */
export function isTruthy (value: unknown): boolean {
  if ([undefined, null, false, 0, ''].includes(value as never)) return false

  return Array.isArray(value) ? value.length > 0 : !isObject(value) || Object.keys(value).length > 0
}

/** A value as `text/template` prints it. */
export function format (value: unknown): string {
  if (value === undefined || value === null) return '<no value>'
  if (Array.isArray(value)) return `[${value.map(item => format(item)).join(' ')}]`
  if (!isObject(value)) return String(value)
  const entries = Object.keys(value).sort(byKey).map(key => `${key}:${format(value[key])}`)

  return `map[${entries.join(' ')}]`
}

function comparable (a: unknown, b: unknown): 'string' | 'number' | 'boolean' {
  const kind = typeof a
  if (typeof b !== kind || !['string', 'number', 'boolean'].includes(kind)) throw new TypeError('incompatible types for comparison')

  return kind as 'string' | 'number' | 'boolean'
}

function ordered (test: (difference: number) => boolean): TemplateFunction {
  return ([a, b]) => {
    if (comparable(a, b) === 'boolean') throw new TypeError('invalid type for comparison')

    return test(a === b ? 0 : (Number((a as string | number) > (b as string | number)) || -1))
  }
}

function indexInto (current: unknown, key: unknown): unknown {
  if (Array.isArray(current)) {
    if (typeof key !== 'number' || key < 0 || key >= current.length) throw new RangeError(`index out of range: ${format(key)}`)

    return current[key]
  }
  if (isObject(current)) return Object.hasOwn(current, String(key)) ? current[String(key)] : undefined
  throw new TypeError(`can't index ${format(current)}`)
}

/** What Sprint puts between operands: a space when neither side is a string. */
function sprint (args: unknown[]): string {
  return args.reduce<string>((out, arg, i) => {
    const space = i > 0 && typeof arg !== 'string' && typeof args[i - 1] !== 'string'

    return `${out}${space ? ' ' : ''}${format(arg)}`
  }, '')
}

/** The functions of `text/template` the subset has. */
export const TEMPLATE_FUNCTIONS: Readonly<Record<string, TemplateFunction>> = {
  eq: ([first, ...others]) => {
    if (others.length === 0) throw new TypeError('missing argument for comparison')

    return others.some((other) => {
      if (first === null || other === null) return first === other
      comparable(first, other)

      return first === other
    })
  },
  ne: ([a, b]) => {
    if (a === null || b === null) return a !== b
    comparable(a, b)

    return a !== b
  },
  lt:  ordered(difference => difference < 0),
  le:  ordered(difference => difference <= 0),
  gt:  ordered(difference => difference > 0),
  ge:  ordered(difference => difference >= 0),
  not: ([value]) => !isTruthy(value),
  and: args => args.find(arg => !isTruthy(arg)) ?? args.at(-1),
  or:  args => args.find(arg => isTruthy(arg)) ?? args.at(-1),
  len: ([value]) => {
    if (typeof value === 'string') return Buffer.byteLength(value)
    if (Array.isArray(value)) return value.length
    if (isObject(value)) return Object.keys(value).length
    throw new TypeError(`len of ${format(value)}`)
  },
  index:  ([value, ...keys]) => keys.reduce<unknown>((current, key) => indexInto(current, key), value),
  print:  args => sprint(args),
  printf: ([pattern, ...args]) => {
    if (typeof pattern !== 'string') throw new TypeError('printf needs a format string')
    let next = 0

    return pattern.replaceAll(/%([%sdvqt])/g, (_match, verb: string) => {
      if (verb === '%') return '%'
      if (next >= args.length) throw new RangeError(`missing argument for %${verb}`)
      const arg = args[next++]

      return verb === 'q' ? JSON.stringify(String(arg)) : format(arg)
    })
  },
}
