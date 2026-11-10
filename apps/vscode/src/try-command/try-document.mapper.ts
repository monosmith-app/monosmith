import { say } from '../words/index.ts'
interface TryCheck { name: string, ok: boolean, tail?: string }

interface TryResult {
  command:      string[]
  commandOk:    boolean
  commandTail?: string
  changed:      string[]
  checks:       TryCheck[]
}

/**
 * Lays out the answer of `smith try <command> --json` as a Markdown page: what
 * the command did in the temporary copy, the files it changed and how each
 * check went.
 * @param json - what smith printed
 * @returns the page, and whether the command and every check passed
 * @throws SyntaxError when the text is not JSON, TypeError when it is not a result
 */
export function toTryDocument (json: string): { page: string, passed: boolean } {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || !Array.isArray((value as TryResult).command)) throw new TypeError(say('vscode.try-command.smith-printed-no-result'))
  const result = value as TryResult
  const checks = result.checks ?? []
  const changed = result.changed ?? []
  const passed = result.commandOk && checks.every((check) => check.ok)
  const tail = (text: string | undefined): string[] => text === undefined || text === '' ? [] : ['', '```', text.trim(), '```']

  return {
    passed,
    page: [
      say('vscode.try-command.try-smith', { join: result.command.join(' ') }),
      '',
      say(passed ? 'vscode.try-command.the-command-and-every-check-passed' : 'vscode.try-command.something-failed-in-the-copy-your'),
      '',
      `- **command:** ${say(result.commandOk ? 'vscode.try-command.ran' : 'vscode.health.failed')}`,
      ...tail(result.commandOk ? undefined : result.commandTail),
      '',
      say('vscode.try-command.files-it-changed'),
      '',
      ...(changed.length === 0 ? [say('vscode.try-command.none')] : changed.map((file) => `- \`${file}\``)),
      '',
      say('vscode.try-command.checks'),
      '',
      ...(checks.length === 0 ? [say('vscode.try-command.none-were-asked-for')] : checks.flatMap((check) => [`- **${check.name}:** ${say(check.ok ? 'vscode.try-command.passed' : 'vscode.health.failed')}`, ...tail(check.ok ? undefined : check.tail)])),
      '',
    ].join('\n'),
  }
}
