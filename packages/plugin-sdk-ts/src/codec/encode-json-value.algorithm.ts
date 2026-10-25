import { EncodeError } from './encode-error.error.ts'

/**
 * A string as the Go SDK writes it: JSON escaping, and `<`, `>`, `&`, U+2028
 * and U+2029 as `\u` escapes (Go's `json.Marshal` escapes HTML by default). Both
 * forms decode to the same text; writing Go's makes the two SDKs' bytes equal.
 */
export function encodeJsonString (text: string): string {
  return JSON.stringify(text)
    .replaceAll('<', String.raw`\u003c`)
    .replaceAll('>', String.raw`\u003e`)
    .replaceAll('&', String.raw`\u0026`)
    .replaceAll('\u{2028}', String.raw`\u2028`)
    .replaceAll('\u{2029}', String.raw`\u2029`)
}

/** Go sorts the keys of a map by their UTF-8 bytes. */
export function compareKeys (a: string, b: string): number {
  return Buffer.compare(Buffer.from(a, 'utf8'), Buffer.from(b, 'utf8'))
}

/**
 * Any JSON value as the Go SDK writes a `map[string]any`: object keys sorted,
 * members whose value is `undefined` left out.
 */
export function encodeJsonValue (value: unknown, path: string): string {
  if (value === null) return 'null'
  switch (typeof value) {
    case 'string': { return encodeJsonString(value)
    }
    case 'boolean': { return value ? 'true' : 'false'
    }
    case 'number': {
      if (!Number.isFinite(value)) throw new EncodeError(`${path}: ${value} cannot be written as JSON`)

      return String(value)
    }
    case 'object': {
      if (Array.isArray(value)) return `[${value.map((item, i) => encodeJsonValue(item === undefined ? null : item, `${path}[${i}]`)).join(',')}]`
      const members = Object.entries(value as Record<string, unknown>)
        .filter(([, member]) => member !== undefined)
        .sort(([a], [b]) => compareKeys(a, b))
        .map(([name, member]) => `${encodeJsonString(name)}:${encodeJsonValue(member, `${path}.${name}`)}`)

      return `{${members.join(',')}}`
    }
    default: {
      throw new EncodeError(`${path}: a ${typeof value} cannot be written as JSON`)
    }
  }
}
