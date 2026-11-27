import { COMMAND_LINES, PALETTE_SECTIONS, type CommandLine, type PaletteSection } from './command-lines.config.ts'

/** A command as `package.json` contributes it, with its title already in words. */
export interface ContributedCommand {
  /** The full id, `monosmith.refresh`. */
  command: string
  title:   string
}

/** One entry of the command list: the command and the smith command line it stands for. */
export interface PaletteEntry {
  command: string
  title:   string
  section: PaletteSection
  /** The smith command line, or undefined for a command of the editor only. */
  line?:   string
}

/** The sections and their entries, in the order the list shows them. */
export interface PaletteGroup {
  section: PaletteSection
  entries: PaletteEntry[]
}

const PREFIX = 'monosmith.'

/**
 * The smith command line a command stands for, as the command list and the views show it.
 * @param command - the full id (`monosmith.buildAll`) or the id without `monosmith.`
 * @param lines - the table of command lines
 * @returns the line, or undefined for a command of the editor only or one the table does not know
 */
export function commandLineOf (command: string, lines: Readonly<Record<string, CommandLine>> = COMMAND_LINES): string | undefined {
  return lines[command.startsWith(PREFIX) ? command.slice(PREFIX.length) : command]?.line
}

/**
 * Groups the commands by section, in the palette's order, each with its smith command line. The
 * commands that are not in the table go last, under Help, so nothing is hidden.
 * @param commands - the commands of `package.json` and the views' own focus commands
 * @param lines - the table of command lines
 * @returns the groups that hold at least one entry, the entries of each by title
 */
export function toPaletteGroups (commands: readonly ContributedCommand[], lines: Readonly<Record<string, CommandLine>> = COMMAND_LINES): PaletteGroup[] {
  const entries = commands.map((command): PaletteEntry => {
    const known = lines[command.command.startsWith(PREFIX) ? command.command.slice(PREFIX.length) : command.command]

    return { command: command.command, title: command.title, section: known?.section ?? 'help', ...(known?.line !== undefined && { line: known.line }) }
  })

  return PALETTE_SECTIONS
    .map((section) => ({ section, entries: entries.filter((entry) => entry.section === section).sort((a, b) => a.title.localeCompare(b.title)) }))
    .filter((group) => group.entries.length > 0)
}
