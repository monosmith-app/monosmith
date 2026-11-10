import type { NeededTools } from './add-options.contract.ts'

/**
 * Reads the tools `smith add` says the new project needs and does not have: the
 * `! <tool> is not installed: ...` lines and the `→ Run smith tools install
 * <names> to choose how to get them.` line it ends them with. smith installs
 * nothing when the editor runs it (no terminal), so this is how the editor
 * learns what to offer.
 * @param lines - what smith printed, one trimmed line each
 * @returns the tools and what is wrong with each, or undefined when none is lacking
 */
export function parseNeededTools (lines: string[]): NeededTools | undefined {
  const command = lines.map((line) => /^→ Run smith tools install (.+?) to choose how to get (?:it|them)\.$/.exec(line)?.[1]).find((names) => names !== undefined)

  return command === undefined
    ? undefined
    : { names: command.split(' '), problems: lines.filter((line) => line.startsWith('! ')).map((line) => line.slice(2)) }
}
