const WORDS: Record<string, string> = {
  added:     'added',
  unchecked: 'added, but a tool is missing, so smith could not check it',
  failing:   'the project was written, and smith doctor found a problem (smith\'s steps follow)',
  written:   'the project was written, and smith doctor found a problem (smith\'s steps follow)',
  composing: 'you chose Compose a stack; nothing was added here',
  cancelled: 'cancelled; nothing was changed',
  copied:    'the command was copied; nothing was changed',
  recipe:    'the recipe was written',
}

/**
 * The line of the Output channel that says how an add or a compose ended, in
 * words that say what happened to the files. "refused" and "cancelled" are not
 * used for a project that was written.
 * @param outcome - how it ended: its kind, and the message of a failure
 * @returns one line, without the name of the command
 */
export function outcomeWords (outcome: { kind: string, message?: string }): string {
  return outcome.kind === 'refused' ? `stopped: ${outcome.message ?? 'smith said no'}` : WORDS[outcome.kind] ?? outcome.kind
}
