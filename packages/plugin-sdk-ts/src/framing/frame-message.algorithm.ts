/**
 * One message as the protocol frames it: `Content-Length: <bytes>`, a blank
 * line (CRLF CRLF), then the body. The length counts bytes of UTF-8, not characters.
 */
export function frameMessage (body: string): Buffer {
  const bytes = Buffer.from(body, 'utf8')

  return Buffer.concat([Buffer.from(`Content-Length: ${bytes.length}\r\n\r\n`, 'ascii'), bytes])
}
