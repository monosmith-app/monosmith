import { tokenizeAction } from './tokenize-action.algorithm.ts'
import type { Token } from './tokenize-action.algorithm.ts'
import type { Command, Operand, Pipeline, TemplateNode } from './template-syntax.contract.ts'
import { TemplateError } from './template.error.ts'

const OPEN = '[['
const CLOSE = ']]'

type Fail = (message: string) => never

/** An `if`, `with` or `range` whose `[[end]]` has not come yet. */
interface Frame {
  type:      'if' | 'with' | 'range'
  pipeline:  Pipeline
  body:      TemplateNode[]
  otherwise: TemplateNode[]
  inElse:    boolean
  /** Created by `else if`: closed together with the frame that holds it. */
  chained:   boolean
}

type Piece = { literal: string } | { action: string }

const UNSUPPORTED = new Set(['define', 'template', 'block', 'break', 'continue'])

/**
 * Parses a template with `[[ ]]` delimiters. The language is the subset of Go's
 * `text/template` that project templates use: `[[.Field]]` and `[[.A.B]]`,
 * `[[if]]`, `[[else if]]`, `[[else]]`, `[[with]]`, `[[range]]` and `[[end]]`, the
 * functions `eq ne lt le gt ge not and or len index print printf`, string,
 * number and boolean literals, parentheses, pipes, comments, and the `-`
 * trim markers (`[[- x -]]`). Variables, `define`, `template` and `block`
 * are not supported and are reported as such.
 */
export function parseTemplate (text: string, name: string): TemplateNode[] {
  const fail: Fail = (message) => {
    throw new TemplateError(name, message)
  }
  const root: TemplateNode[] = []
  const stack: Frame[] = []
  const target = (): TemplateNode[] => {
    const top = stack.at(-1)
    if (top === undefined) return root

    return top.inElse ? top.otherwise : top.body
  }

  for (const piece of splitPieces(text, fail)) {
    if ('literal' in piece) {
      if (piece.literal !== '') target().push({ type: 'text', text: piece.literal })
    } else if (piece.action.startsWith('/*')) {
      if (!piece.action.endsWith('*/')) fail('unclosed comment')
    } else {
      applyAction(piece.action, stack, target, fail)
    }
  }
  const open = stack.at(-1)
  if (open !== undefined) fail(`unclosed ${open.type}: missing [[end]]`)

  return root
}

/** The literal text and the actions of a template, in order, with the `-` trim markers applied. */
function splitPieces (text: string, fail: Fail): Piece[] {
  const pieces: Piece[] = []
  let position = 0
  let trimNext = false
  while (position < text.length) {
    const open = text.indexOf(OPEN, position)
    let literal = text.slice(position, open === -1 ? text.length : open)
    if (trimNext) literal = literal.trimStart()
    trimNext = false
    if (open === -1) {
      pieces.push({ literal })
      break
    }
    const close = findClose(text, open + OPEN.length)
    if (close === -1) fail(`unclosed action at offset ${open}`)
    let action = text.slice(open + OPEN.length, close)
    if (/^-\s/.test(action)) {
      literal = literal.trimEnd()
      action = action.slice(1)
    }
    if (/\s-$/.test(action)) {
      trimNext = true
      action = action.slice(0, -1)
    }
    pieces.push({ literal }, { action: action.trim() })
    position = close + CLOSE.length
  }

  return pieces
}

/** The offset of the `]]` that ends the action, skipping string literals. */
function findClose (text: string, from: number): number {
  for (let i = from; i < text.length; i++) {
    const c = text[i]
    if (c === '"') {
      i++
      while (i < text.length && text[i] !== '"') i += text[i] === '\\' ? 2 : 1
    } else if (c === '`') {
      i = text.indexOf('`', i + 1)
      if (i === -1) return -1
    } else if (text.startsWith(CLOSE, i)) {
      return i
    }
  }

  return -1
}

function applyAction (action: string, stack: Frame[], target: () => TemplateNode[], fail: Fail): void {
  const tokens = tokenizeAction(action, fail)
  const first = tokens[0]
  const keyword = first?.t === 'word' ? first.value : ''
  switch (keyword) {
    case 'if':
    case 'with':
    case 'range': {
      stack.push({ type: keyword, pipeline: parsePipeline(tokens.slice(1), fail), body: [], otherwise: [], inElse: false, chained: false })

      break
    }
    case 'else': {
      startElse(tokens, stack, fail)

      break
    }
    case 'end': {
      if (tokens.length > 1) fail('unexpected tokens after end')
      closeFrames(stack, target, fail)

      break
    }
    default: { if (UNSUPPORTED.has(keyword)) {
      fail(`"${keyword}" is not supported by the SDK's templates (the supported subset: fields, if, else, with, range, end, and the functions eq ne lt le gt ge not and or len index print printf)`)
    } else {
      if (tokens.length === 0) fail('empty action')
      target().push({ type: 'action', pipeline: parsePipeline(tokens, fail) })
    }
    }
  }
}

function startElse (tokens: Token[], stack: Frame[], fail: Fail): void {
  const top = stack.at(-1) ?? fail('unexpected else')
  if (top.inElse) fail('multiple else in one block')
  top.inElse = true
  const next = tokens[1]
  if (next?.t === 'word' && next.value === 'if') {
    stack.push({ type: 'if', pipeline: parsePipeline(tokens.slice(2), fail), body: [], otherwise: [], inElse: false, chained: true })
  } else if (tokens.length > 1) {
    fail('unexpected tokens after else')
  }
}

/** Ends the innermost block, and the `else if` blocks chained to it. */
function closeFrames (stack: Frame[], target: () => TemplateNode[], fail: Fail): void {
  let frame: Frame
  do {
    frame = stack.pop() ?? fail('unexpected end')
    target().push({ type: frame.type, pipeline: frame.pipeline, body: frame.body, otherwise: frame.otherwise })
  } while (frame.chained)
}

function parsePipeline (tokens: Token[], fail: Fail): Pipeline {
  const [pipeline, used] = pipelineFrom(tokens, 0, fail)
  if (used < tokens.length) fail('unexpected ")" in action')

  return pipeline
}

function pipelineFrom (tokens: Token[], start: number, fail: Fail): [Pipeline, number] {
  const pipeline: Pipeline = []
  let command: Command = []
  let i = start
  for (; i < tokens.length; i++) {
    const token = tokens[i]
    if (token.t === 'close') break
    if (token.t === 'pipe') {
      if (command.length === 0) fail('missing value before |')
      pipeline.push(command)
      command = []
    } else if (token.t === 'open') {
      const [inner, next] = pipelineFrom(tokens, i + 1, fail)
      if (tokens[next]?.t !== 'close') fail('unclosed (')
      command.push({ kind: 'pipeline', pipeline: inner })
      i = next
    } else {
      command.push(operandOf(token))
    }
  }
  if (command.length === 0) fail(pipeline.length > 0 ? 'missing command after |' : 'missing value for the action')
  pipeline.push(command)

  return [pipeline, i]
}

function operandOf (token: Token): Operand {
  if (token.t === 'field') return { kind: 'field', root: token.root, path: token.path }
  if (token.t === 'string' || token.t === 'number') return { kind: token.t, value: token.value }
  if (token.t !== 'word') throw new Error('unreachable')
  if (token.value === 'true' || token.value === 'false') return { kind: 'bool', value: token.value === 'true' }

  return token.value === 'nil' ? { kind: 'nil' } : { kind: 'name', name: token.value }
}
