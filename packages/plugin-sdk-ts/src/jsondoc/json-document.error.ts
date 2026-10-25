/** The text is not a JSON object, or not JSON at all. */
export class JsonDocumentError extends Error {
  constructor (message: string, options?: ErrorOptions) {
    super(message, options)
    this.name = 'JsonDocumentError'
  }
}
