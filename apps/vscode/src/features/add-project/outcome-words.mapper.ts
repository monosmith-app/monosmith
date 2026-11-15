import { say } from '../words/index.ts'
const WORDS: Record<string, string> = {
  added:     'added',
  unchecked: say('vscode.add-project.added-but-a-tool-is-missing'),
  failing:   say('vscode.add-project.the-project-was-written-and-smith'),
  written:   say('vscode.add-project.the-project-was-written-and-smith'),
  composing: say('vscode.add-project.you-chose-compose-a-stack-nothing'),
  cancelled: say('vscode.add-project.cancelled-nothing-was-changed'),
  copied:    say('vscode.add-project.the-command-was-copied-nothing-was'),
  recipe:    say('vscode.add-project.the-recipe-was-written'),
}

/**
 * The line of the Output channel that says how an add or a compose ended, in
 * words that say what happened to the files. "refused" and "cancelled" are not
 * used for a project that was written.
 * @param outcome - how it ended: its kind, and the message of a failure
 * @returns one line, without the name of the command
 */
export function outcomeWords (outcome: { kind: string, message?: string }): string {
  return outcome.kind === 'refused' ? `stopped: ${outcome.message ?? say('vscode.add-project.smith-said-no')}` : WORDS[outcome.kind] ?? outcome.kind
}
