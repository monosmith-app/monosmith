import { MAX_MESSAGE_BYTES } from '../protocol/index.ts'
import { FramingError } from './framing.error.ts'

/**
 * Cuts a byte stream into the bodies of its frames. Bytes arrive in chunks of
 * any size; {@link FrameReader.push} returns every body a chunk completed.
 * Headers end at a blank line (CRLF or LF alone, as the Go SDK reads them),
 * `Content-Length` is matched without regard to case, and a header that is not
 * `Name: value`, a length that is not a number from 0 to 64 MiB, or a header
 * block without a length is a {@link FramingError}.
 */
export class FrameReader {
  #buffer: Buffer = Buffer.alloc(0)

  /** The length and where the body starts, or null while the header is incomplete. */
  #readHeader (): { length: number, bodyStart: number } | null {
    let length = -1
    let position = 0
    for (;;) {
      const newline = this.#buffer.indexOf(0x0A, position)
      if (newline === -1) return null
      const line = this.#buffer.subarray(position, newline).toString('ascii').replace(/[\r\n]+$/, '')
      position = newline + 1
      if (line === '') break
      const colon = line.indexOf(':')
      if (colon === -1) throw new FramingError(`malformed header "${line}"`)
      if (line.slice(0, colon).trim().toLowerCase() !== 'content-length') continue
      const text = line.slice(colon + 1).trim()
      const parsed = /^\d+$/.test(text) ? Number(text) : NaN
      if (!Number.isSafeInteger(parsed) || parsed > MAX_MESSAGE_BYTES) throw new FramingError(`invalid Content-Length "${line.slice(colon + 1)}"`)
      length = parsed
    }
    if (length < 0) throw new FramingError('message without Content-Length')

    return { length, bodyStart: position }
  }

  /** Bytes received that are not part of a complete frame yet. */
  get pending (): number {
    return this.#buffer.length
  }

  /** Adds the bytes that arrived and returns the body of every frame they completed, in order. */
  push (chunk: Uint8Array): string[] {
    this.#buffer = this.#buffer.length === 0 ? Buffer.from(chunk) : Buffer.concat([this.#buffer, chunk])
    const bodies: string[] = []
    for (;;) {
      const header = this.#readHeader()
      if (header === null) return bodies
      const end = header.bodyStart + header.length
      if (this.#buffer.length < end) return bodies
      bodies.push(this.#buffer.subarray(header.bodyStart, end).toString('utf8'))
      this.#buffer = this.#buffer.subarray(end)
    }
  }
}
