/** A frame whose body is not a JSON-RPC 2.0 message of the shape the protocol uses. The serve loop answers it with the code ParseError. */
export class MessageParseError extends Error {
  constructor (message: string) {
    super(message)
    this.name = 'MessageParseError'
  }
}
