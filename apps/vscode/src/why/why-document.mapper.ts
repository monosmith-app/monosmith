import { say } from '../words/index.ts'
type Answer = Record<string, unknown>

const FILE_ORDER = ['owner', 'kind', 'project', 'rule', 'renderer', 'plugin', 'source', 'state']
/** Keys the page says in its own place, so the list of facts leaves them out. */
const SAID_ELSEWHERE = new Set(['path', 'known', 'summary', 'change', 'task', 'status', 'key', 'lastKey', 'lastRun', 'differences'])

function asAnswer (json: string): Answer {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new TypeError(say('vscode.why.smith-printed-no-answer'))

  return value as Answer
}

function show (value: unknown): string {
  return typeof value === 'string' ? value : JSON.stringify(value)
}

function facts (answer: Answer, order: readonly string[]): string[] {
  const keys = [...order.filter((key) => answer[key] !== undefined), ...Object.keys(answer).filter((key) => !order.includes(key) && !SAID_ELSEWHERE.has(key))]

  return keys.filter((key) => !SAID_ELSEWHERE.has(key) && answer[key] !== undefined && answer[key] !== '').map((key) => `- **${key}:** ${show(answer[key])}`)
}

/**
 * Lays out the answer of `smith why <file> --json` as a Markdown page: who
 * owns the file, what produced it, and how to change it.
 * @param json - what smith printed
 * @returns the page
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function toFileWhyDocument (json: string): string {
  const answer = asAnswer(json)
  const change = Array.isArray(answer.change) ? answer.change.map((step) => `- ${show(step)}`) : []
  const known = answer.known === false

  return [
    say('vscode.why.why', { show: show(answer.path ?? say('vscode.why.this-file')) }),
    '',
    show(answer.summary ?? say('vscode.why.smith-gave-no-summary')),
    '',
    ...(known ? [] : [...facts(answer, FILE_ORDER), '']),
    say('vscode.why.how-to-change-it'),
    '',
    ...(change.length === 0 ? [say('vscode.why.smith-named-no-step')] : change),
    '',
  ].join('\n')
}

/**
 * Lays out the answer of `smith why <setting> --json` as a Markdown page: what
 * the setting does, its value and where it comes from, why its default is what
 * it is, and how to change it and go back.
 * @param json - what smith printed
 * @returns the page
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function toSettingWhyDocument (json: string): string {
  const answer = asAnswer(json)
  const set = answer.source === 'smith.toml'
  const value = set ? show(answer.value) : say('vscode.why.the-default-smith-toml-does-not', { show: show(answer.default ?? 'none') })

  return [
    say('vscode.why.why', { show: show(answer.path ?? say('vscode.why.this-setting')) }),
    '',
    show(answer.explanation ?? say('vscode.why.smith-gave-no-explanation')),
    '',
    `- **value:** ${value}`,
    ...(set && answer.default !== undefined ? [`- **default:** ${show(answer.default)}`] : []),
    ...(answer.reason === undefined || answer.reason === '' ? [] : ['', say('vscode.why.why-the-default-is-what-it'), '', show(answer.reason)]),
    '',
    say('vscode.why.how-to-change-it'),
    '',
    `- ${show(answer.change ?? say('vscode.why.smith-named-no-step'))}`,
    ...(answer.undo === undefined || answer.undo === '' ? [] : [`- to go back: ${show(answer.undo)}`]),
    '',
  ].join('\n')
}

/**
 * Lays out the answer of `smith why-miss <project>:<task> --json` as a
 * Markdown page: whether the task would run again, and what differs from the
 * last run.
 * @param json - what smith printed
 * @returns the page
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function toTaskWhyDocument (json: string): string {
  const answer = asAnswer(json)
  const status = show(answer.status ?? 'unknown')
  const differences = Array.isArray(answer.differences) ? answer.differences as Answer[] : []
  const summary: Record<string, string> = {
    'hit':            say('vscode.why.nothing-changed-since-the-last-run'),
    'changed':        say('vscode.why.the-task-runs-again-because-its'),
    'no-earlier-run': say('vscode.why.the-task-has-no-earlier-run'),
  }

  return [
    say('vscode.why.why', { show: show(answer.task ?? say('vscode.why.this-task')) }),
    '',
    summary[status] ?? say('vscode.why.smith-reports-the-state', { status }),
    '',
    `- **state:** ${status}`,
    ...(answer.lastRun === undefined ? [] : [say('vscode.why.last-run', { show: show(answer.lastRun) })]),
    ...(differences.length === 0 ? [] : ['', say('vscode.why.what-changed'), '', ...differences.map((item) => `- ${show(item.what)} \`${show(item.name)}\`: ${show(item.change)}`)]),
    '',
    say('vscode.why.run-monosmith-explain-project-to-see'),
    '',
  ].join('\n')
}
