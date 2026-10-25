/** The bytes on the stream are not a sequence of frames: a malformed header, a missing or invalid Content-Length. The stream cannot be trusted after it. */
export class FramingError extends Error {
  constructor (message: string) {
    super(message)
    this.name = 'FramingError'
  }
}
