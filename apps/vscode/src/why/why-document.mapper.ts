type Answer = Record<string, unknown>

const FILE_ORDER = ['owner', 'kind', 'project', 'rule', 'renderer', 'plugin', 'source', 'state']
/** Keys the page says in its own place, so the list of facts leaves them out. */
const SAID_ELSEWHERE = new Set(['path', 'known', 'summary', 'change', 'task', 'status', 'key', 'lastKey', 'lastRun', 'differences'])

function asAnswer (json: string): Answer {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new TypeError('smith printed no answer')

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
    `# Why: ${show(answer.path ?? 'this file')}`,
    '',
    show(answer.summary ?? 'smith gave no summary.'),
    '',
    ...(known ? [] : [...facts(answer, FILE_ORDER), '']),
    '## How to change it',
    '',
    ...(change.length === 0 ? ['smith named no step.'] : change),
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
  const value = set ? show(answer.value) : `${show(answer.default ?? 'none')} (the default: smith.toml does not set it)`

  return [
    `# Why: ${show(answer.path ?? 'this setting')}`,
    '',
    show(answer.explanation ?? 'smith gave no explanation.'),
    '',
    `- **value:** ${value}`,
    ...(set && answer.default !== undefined ? [`- **default:** ${show(answer.default)}`] : []),
    ...(answer.reason === undefined || answer.reason === '' ? [] : ['', '## Why the default is what it is', '', show(answer.reason)]),
    '',
    '## How to change it',
    '',
    `- ${show(answer.change ?? 'smith named no step.')}`,
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
    'hit':            'Nothing changed since the last run, so the task comes from the cache.',
    'changed':        'The task runs again because its inputs changed since the last run.',
    'no-earlier-run': 'The task has no earlier run in the cache, so it runs.',
  }

  return [
    `# Why: ${show(answer.task ?? 'this task')}`,
    '',
    summary[status] ?? `smith reports the state "${status}".`,
    '',
    `- **state:** ${status}`,
    ...(answer.lastRun === undefined ? [] : [`- **last run:** ${show(answer.lastRun)}`]),
    ...(differences.length === 0 ? [] : ['', '## What changed', '', ...differences.map((item) => `- ${show(item.what)} \`${show(item.name)}\`: ${show(item.change)}`)]),
    '',
    'Run `MonoSmith: Explain Project` to see where each setting of the task comes from.',
    '',
  ].join('\n')
}
