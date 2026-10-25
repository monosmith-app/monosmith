/** One operand of a command: a field of the data, a literal, a parenthesised pipeline or a function name. */
export type Operand =
  | { kind: 'field', root: boolean, path: string[] } |
  { kind: 'string' | 'number' | 'bool', value: string | number | boolean } |
  { kind: 'nil' } |
  { kind: 'pipeline', pipeline: Pipeline } |
  { kind: 'name', name: string }

/** Operands in a row: a function call (`eq .Kind "app"`) or a single value. */
export type Command = Operand[]

/** Commands joined by `|`, each receiving the previous one's value as its last argument. */
export type Pipeline = Command[]

export type TemplateNode =
  | { type: 'text', text: string } |
  { type: 'action', pipeline: Pipeline } |
  { type: 'if' | 'with' | 'range', pipeline: Pipeline, body: TemplateNode[], otherwise: TemplateNode[] }
