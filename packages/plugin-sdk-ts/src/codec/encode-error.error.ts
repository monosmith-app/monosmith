/** A value that cannot be written as the message the schema describes: a required member is missing, or a member has the wrong type. */
export class EncodeError extends Error {
  constructor (message: string) {
    super(message)
    this.name = 'EncodeError'
  }
}
