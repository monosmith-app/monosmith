import { parseTemplate } from './parse-template.algorithm.ts'
import { TEMPLATE_FUNCTIONS, byKey, format, isObject, isTruthy } from './template-functions.algorithm.ts'
import type { Command, Operand, Pipeline, TemplateNode } from './template-syntax.contract.ts'
import { TemplateError } from './template.error.ts'

interface Scope {
  dot:  unknown
  root: unknown
}

type Fail = (message: string) => never

/**
 * Renders a `[[ ]]` template with the data, the way the Go SDK's `templating`
 * does with `text/template` and `missingkey=error`: a field the data does not
 * have is an error, not an empty string. See {@link parseTemplate} for the
 * language. `name` is only used in error messages.
 */
export function renderTemplate (text: string, data: unknown, name = 'template'): string {
  const fail: Fail = (message) => {
    throw new TemplateError(name, message)
  }

  return render(parseTemplate(text, name), { dot: data, root: data }, fail)
}

function render (nodes: TemplateNode[], scope: Scope, fail: Fail): string {
  return nodes.map(node => renderNode(node, scope, fail)).join('')
}

function renderNode (node: TemplateNode, scope: Scope, fail: Fail): string {
  if (node.type === 'text') return node.text
  const value = evaluatePipeline(node.pipeline, scope, fail)
  if (node.type === 'action') return format(value)
  if (node.type === 'if') return render(isTruthy(value) ? node.body : node.otherwise, scope, fail)
  if (node.type === 'with') return isTruthy(value) ? render(node.body, { dot: value, root: scope.root }, fail) : render(node.otherwise, scope, fail)
  const items = rangeItems(value, fail)

  return items.length === 0 ? render(node.otherwise, scope, fail) : items.map(item => render(node.body, { dot: item, root: scope.root }, fail)).join('')
}

/** What `range` loops over: the elements of an array, or the values of a map in key order. */
function rangeItems (value: unknown, fail: Fail): unknown[] {
  if (Array.isArray(value)) return value
  if (value === null || value === undefined) return []

  return isObject(value) ? Object.keys(value).sort(byKey).map(key => value[key]) : fail(`range can't iterate over ${format(value)}`)
}

const NONE = Symbol('nothing piped')

function evaluatePipeline (pipeline: Pipeline, scope: Scope, fail: Fail): unknown {
  let piped: unknown = NONE
  for (const command of pipeline) piped = evaluateCommand(command, scope, piped, fail)

  return piped
}

function evaluateCommand (command: Command, scope: Scope, piped: unknown, fail: Fail): unknown {
  const [head, ...rest] = command as [Operand, ...Operand[]]
  if (head.kind === 'name') return callFunction(head.name, rest, scope, piped, fail)
  if (rest.length > 0) return fail('can\'t give an argument to a value that is not a function')

  return piped === NONE ? evaluateOperand(head, scope, fail) : fail('can\'t pipe into a value that is not a function')
}

function callFunction (name: string, operands: Operand[], scope: Scope, piped: unknown, fail: Fail): unknown {
  const fn = Object.hasOwn(TEMPLATE_FUNCTIONS, name) ? TEMPLATE_FUNCTIONS[name] : undefined
  if (fn === undefined) return fail(`function "${name}" not defined`)
  const args = operands.map(operand => evaluateOperand(operand, scope, fail))
  if (piped !== NONE) args.push(piped)
  try {
    return fn(args)
  } catch (error) {
    return fail(`error calling ${name}: ${error instanceof Error ? error.message : String(error)}`)
  }
}

function evaluateOperand (operand: Operand, scope: Scope, fail: Fail): unknown {
  if (operand.kind === 'nil') return null
  if ('value' in operand) return operand.value
  if (operand.kind === 'pipeline') return evaluatePipeline(operand.pipeline, scope, fail)
  if (operand.kind === 'name') return evaluateCommand([operand], scope, NONE, fail)

  let value = operand.root ? scope.root : scope.dot
  for (const key of operand.path) {
    if (!isObject(value)) return fail(`can't evaluate field ${key} in ${value === null || value === undefined ? 'nil' : typeof value}`)
    if (!Object.hasOwn(value, key)) return fail(`map has no entry for key "${key}"`)
    value = value[key]
  }

  return value
}
