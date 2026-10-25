import { describe, expect, it } from 'vitest'
import { frameMessage } from './frame-message.algorithm.ts'
import { FrameReader } from './frame-reader.store.ts'
import { FramingError } from './framing.error.ts'

describe('frameMessage', () => {
  it('counts bytes, not characters', () => {
    const frame = frameMessage('{"a":"é€"}')
    expect(frame.toString('utf8')).toBe('Content-Length: 13\r\n\r\n{"a":"é€"}')
  })
})

describe('FrameReader', () => {
  it('cuts frames that arrive in one chunk', () => {
    const reader = new FrameReader()
    const frames = Buffer.concat([frameMessage('{"a":1}'), frameMessage('{"b":2}')])
    expect(reader.push(frames)).toEqual(['{"a":1}', '{"b":2}'])
    expect(reader.pending).toBe(0)
  })

  it('waits for the rest when a frame arrives byte by byte, even inside a multi-byte character', () => {
    const reader = new FrameReader()
    const bytes = frameMessage('{"a":"é€"}')
    const bodies: string[] = []
    for (const byte of bytes) bodies.push(...reader.push(Uint8Array.of(byte)))
    expect(bodies).toEqual(['{"a":"é€"}'])
  })

  it('matches Content-Length without regard to case, accepts bare LF line ends and other headers', () => {
    const reader = new FrameReader()
    expect(reader.push(Buffer.from('content-length:  2\nX-Other: yes\n\n{}'))).toEqual(['{}'])
  })

  it('refuses a malformed header, a missing or invalid length, and a length beyond 64 MiB', () => {
    expect(() => new FrameReader().push(Buffer.from('hello world\r\n\r\n'))).toThrow(new FramingError('malformed header "hello world"'))
    expect(() => new FrameReader().push(Buffer.from('X: 1\r\n\r\n'))).toThrow('message without Content-Length')
    expect(() => new FrameReader().push(Buffer.from('Content-Length: abc\r\n\r\n'))).toThrow('invalid Content-Length')
    expect(() => new FrameReader().push(Buffer.from('Content-Length: -1\r\n\r\n'))).toThrow('invalid Content-Length')
    expect(() => new FrameReader().push(Buffer.from(`Content-Length: ${64 * 1024 * 1024 + 1}\r\n\r\n`))).toThrow('invalid Content-Length')
  })
})
