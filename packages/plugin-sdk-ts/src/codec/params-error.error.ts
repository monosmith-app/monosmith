/** A received message whose params (or result) do not have the shape the schema describes. The serve loop answers it with the code InvalidParams. */
export class ParamsError extends Error {
  constructor (message: string) {
    super(message)
    this.name = 'ParamsError'
  }
}
