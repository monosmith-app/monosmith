/** A template that cannot be parsed or rendered: bad syntax, a field the data does not have, a construct outside the supported subset. */
export class TemplateError extends Error {
  constructor (name: string, message: string) {
    super(`template ${name}: ${message}`)
    this.name = 'TemplateError'
  }
}
