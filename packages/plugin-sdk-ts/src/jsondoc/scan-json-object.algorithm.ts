import { JsonDocumentError } from './json-document.error.ts'

/** A JSON value as the document keeps it: an object edited member by member, anything else as the text it was written in. */
export type JsonNode =
  | { kind: 'object', members: Array<[string, JsonNode]> } |
  { kind: 'raw', text: string }

const WHITESPACE = new Set([' ', '\t', '\n', '\r'])

/**
 * Reads the text of a JSON object into a tree that keeps the order of its
 * members and the exact text of every leaf and array (`1.0` stays `1.0`). A
 * duplicate key keeps its first position and takes its last value, as the Go
 * SDK's document does.
 */
export function scanJsonObject (text: string): Extract<JsonNode, { kind: 'object' }> {
  const scanner = new Scanner(text)
  scanner.skip()
  if (text[scanner.position] !== '{') throw new JsonDocumentError('not a JSON object')
  const root = scanner.object()
  scanner.skip()
  if (scanner.position < text.length) throw new JsonDocumentError(`unexpected "${text[scanner.position]}" after the object, at ${scanner.position}`)

  return root
}

class Scanner {
  readonly #text: string
  position = 0

  constructor (text: string) {
    this.#text = text
  }

  skip (): void {
    while (this.position < this.#text.length && WHITESPACE.has(this.#text[this.position])) this.position++
  }

  object (): Extract<JsonNode, { kind: 'object' }> {
    const members: Array<[string, JsonNode]> = []
    const index = new Map<string, number>()
    this.position++ // {
    this.skip()
    if (this.#text[this.position] === '}') {
      this.position++

      return { kind: 'object', members }
    }
    for (;;) {
      this.skip()
      if (this.#text[this.position] !== '"') throw this.fail('an object key')
      const key = JSON.parse(this.string()) as string
      this.skip()
      if (this.#text[this.position] !== ':') throw this.fail('":"')
      this.position++
      const value = this.value()
      const at = index.get(key)
      if (at === undefined) {
        index.set(key, members.length)
        members.push([key, value])
      } else {
        members[at] = [key, value]
      }
      this.skip()
      const next = this.#text[this.position]
      if (next === '}') {
        this.position++

        return { kind: 'object', members }
      }
      if (next !== ',') throw this.fail('"," or "}"')
      this.position++
    }
  }

  value (): JsonNode {
    this.skip()
    const first = this.#text[this.position]
    if (first === '{') return this.object()
    const start = this.position
    if (first === '[') this.array()
    else if (first === '"') this.string()
    else this.literal()
    const text = this.#text.slice(start, this.position)
    try {
      JSON.parse(text)
    } catch (error) {
      throw new JsonDocumentError(`invalid JSON value "${text.slice(0, 40)}" at ${start}`, { cause: error })
    }

    return { kind: 'raw', text }
  }

  /** Moves past an array, whose elements stay text. */
  array (): void {
    let depth = 0
    while (this.position < this.#text.length) {
      const c = this.#text[this.position]
      if (c === '"') {
        this.string()
        continue
      }
      if (c === '[' || c === '{') depth++
      else if (c === ']' || c === '}') depth--
      this.position++
      if (depth === 0) return
    }
    throw new JsonDocumentError('unterminated array')
  }

  /** Moves past a string literal and returns its text, quotes included. */
  string (): string {
    const start = this.position
    this.position++
    while (this.position < this.#text.length) {
      const c = this.#text[this.position++]
      if (c === '\\') this.position++
      else if (c === '"') return this.#text.slice(start, this.position)
    }
    throw new JsonDocumentError('unterminated string')
  }

  literal (): void {
    const start = this.position
    while (this.position < this.#text.length && !WHITESPACE.has(this.#text[this.position]) && !',}]'.includes(this.#text[this.position])) this.position++
    if (this.position === start) throw this.fail('a value')
  }

  fail (expected: string): JsonDocumentError {
    const found = this.position < this.#text.length ? `"${this.#text[this.position]}"` : 'the end of the text'

    return new JsonDocumentError(`expected ${expected}, found ${found} at ${this.position}`)
  }
}
