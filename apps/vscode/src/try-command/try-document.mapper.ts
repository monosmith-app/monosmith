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
  if (typeof value !== 'object' || value === null || !Array.isArray((value as TryResult).command)) throw new TypeError('smith printed no result')
  const result = value as TryResult
  const checks = result.checks ?? []
  const changed = result.changed ?? []
  const passed = result.commandOk && checks.every((check) => check.ok)
  const tail = (text: string | undefined): string[] => text === undefined || text === '' ? [] : ['', '```', text.trim(), '```']

  return {
    passed,
    page: [
      `# Try: smith ${result.command.join(' ')}`,
      '',
      passed ? 'The command and every check passed in the copy. Your files were not touched.' : 'Something failed in the copy. Your files were not touched.',
      '',
      `- **command:** ${result.commandOk ? 'ran' : 'failed'}`,
      ...tail(result.commandOk ? undefined : result.commandTail),
      '',
      '## Files it changed',
      '',
      ...(changed.length === 0 ? ['None.'] : changed.map((file) => `- \`${file}\``)),
      '',
      '## Checks',
      '',
      ...(checks.length === 0 ? ['None were asked for.'] : checks.flatMap((check) => [`- **${check.name}:** ${check.ok ? 'passed' : 'failed'}`, ...tail(check.ok ? undefined : check.tail)])),
      '',
    ].join('\n'),
  }
}
