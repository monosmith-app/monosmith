import { indentRawJson } from './indent-raw-json.algorithm.ts'
import { JsonDocumentError } from './json-document.error.ts'
import { scanJsonObject } from './scan-json-object.algorithm.ts'
import type { JsonNode } from './scan-json-object.algorithm.ts'

type Entry = JsonDocument | { raw: string }

/** A string as Go's JSON encoder writes it without HTML escaping: JSON, with U+2028 and U+2029 escaped. */
function quote (text: string): string {
  return JSON.stringify(text)
    .replaceAll('\u{2028}', String.raw`\u2028`)
    .replaceAll('\u{2029}', String.raw`\u2029`)
}

/**
 * A JSON object that keeps its key order and the text of everything you do not
 * touch, so editing a `package.json` changes only what was edited: set a
 * member and it keeps its place, add one and it goes last, delete one and the
 * others stay where they were. Port of the Go SDK's `jsondoc.Document`.
 *
 * ```ts
 * const doc = JsonDocument.parse(text)
 * doc.setString('version', '1.2.3')
 * writeFileSync(path, doc.render())
 * ```
 */
export class JsonDocument {
  /** Reads a JSON object. Throws {@link JsonDocumentError} for anything else. */
  static parse (text: string): JsonDocument {
    return JsonDocument.#fromNode(scanJsonObject(text))
  }

  static #fromNode (node: Extract<JsonNode, { kind: 'object' }>): JsonDocument {
    const doc = new JsonDocument()
    for (const [key, value] of node.members) doc.#entries.set(key, value.kind === 'object' ? JsonDocument.#fromNode(value) : { raw: value.text })

    return doc
  }

  /** In the order the keys were first set: a Map keeps it, whatever the keys look like. */
  readonly #entries = new Map<string, Entry>()

  #write (indent: string): string {
    if (this.#entries.size === 0) return '{}'
    const inner = `${indent}  `
    const lines = [...this.#entries].map(([key, entry]) => {
      const value = entry instanceof JsonDocument ? entry.#write(inner) : indentRawJson(entry.raw, inner)

      return `${inner}${quote(key)}: ${value}`
    })

    return `{\n${lines.join(',\n')}\n${indent}}`
  }

  /** Whether the key exists. */
  has (key: string): boolean {
    return this.#entries.has(key)
  }

  /** The keys, in order. */
  keys (): string[] {
    // Iterator#toArray needs Node 22; the SDK runs on 20.
    // eslint-disable-next-line unicorn/prefer-spread
    return Array.from(this.#entries.keys())
  }

  /** The number of keys. */
  get size (): number {
    return this.#entries.size
  }

  /** The object under a key, editable in place; `create` adds an empty one (last) when there is none. Undefined for a key that holds something else. */
  object (key: string, create = false): JsonDocument | undefined {
    const entry = this.#entries.get(key)
    if (entry instanceof JsonDocument) return entry
    if (!create) return undefined
    const child = new JsonDocument()
    this.#entries.set(key, child)

    return child
  }

  /** A string value; undefined when the key is missing or does not hold a string. */
  getString (key: string): string | undefined {
    const entry = this.#entries.get(key)
    if (entry === undefined || entry instanceof JsonDocument) return undefined
    const value: unknown = JSON.parse(entry.raw)

    return typeof value === 'string' ? value : undefined
  }

  /** Any value under a key as JavaScript sees it, or undefined for a missing key. */
  getValue (key: string): unknown {
    const entry = this.#entries.get(key)
    if (entry === undefined) return undefined

    return JSON.parse(entry instanceof JsonDocument ? entry.render() : entry.raw)
  }

  /** Sets a string, keeping the key's position. */
  setString (key: string, value: string): void {
    this.#entries.set(key, { raw: quote(value) })
  }

  /** Sets any JSON-encodable value, keeping the key's position. A plain object becomes an editable document, with its members in the order you wrote them. */
  setJson (key: string, value: unknown): void {
    if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      this.#entries.set(key, JsonDocument.parse(JSON.stringify(value)))

      return
    }
    const raw = JSON.stringify(value)
    if (raw === undefined) throw new JsonDocumentError(`${key}: a ${typeof value} cannot be written as JSON`)
    this.#entries.set(key, { raw })
  }

  /** Removes a key. */
  delete (key: string): void {
    this.#entries.delete(key)
  }

  /** The document as npm writes a package.json: two-space indentation, a final newline. */
  render (): string {
    return `${this.#write('')}\n`
  }

  toString (): string {
    return this.render()
  }
}
