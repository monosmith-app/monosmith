/**
 * The arguments of `smith convert` for the answers to its questions. The four choices apply to the
 * full level only, as smith's own `when` says, so a conservative conversion names none.
 * @param answers - the answers by question key (`level`, `lint`, `test`, `build`, `ci`)
 * @param apply - false is the plan (`--dry-run`); true writes (`--yes`)
 * @param commit - add `--commit`, so one commit that `smith undo` reverts (only when applying)
 * @returns the arguments after `smith`
 */
export function toConvertArguments (answers: Readonly<Record<string, string>>, apply: boolean, commit = false): string[] {
  const level = answers.level ?? 'conservative'

  return [
    'convert', '--level', level,
    ...(level === 'full' ? ['--lint', answers.lint ?? 'keep', '--test', answers.test ?? 'keep', '--build', answers.build ?? 'smith', '--ci', answers.ci ?? 'smith'] : []),
    ...(apply ? ['--yes', ...(commit ? ['--commit'] : [])] : ['--dry-run']),
  ]
}
