/** What answering one request came to: the result, already written as JSON, or an error object. */
export type RequestOutcome =
  | { ok: true, resultJson: string } |
  { ok: false, code: number, message: string }
