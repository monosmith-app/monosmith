import * as vscode from 'vscode'
import type { PaletteSection } from './command-lines.config.ts'
import { toPaletteGroups, type ContributedCommand } from './palette-entries.mapper.ts'
import { say, type WordKey } from '../../shared/words/index.ts'

const SECTION_WORDS: Record<PaletteSection, WordKey> = {
  start:        'commands.section.start',
  build:        'commands.section.build',
  projects:     'commands.section.projects',
  dependencies: 'commands.section.dependencies',
  release:      'commands.section.release',
  health:       'commands.section.health',
  goto:         'commands.section.goto',
  help:         'commands.section.help',
}

/** A view of the MonoSmith container, for the Go to entries. */
export interface PaletteView {
  /** The view id, `monosmith.home`. */
  id:   string
  name: string
}

/** The title of a command of `package.json`: its `%key%` read from the catalogue. */
function titleOf (title: string): string {
  const key = /^%(.+)%$/.exec(title)?.[1]

  return key === undefined ? title : say(key as WordKey)
}

/**
 * Shows every command of the extension in one list, grouped as the terminal interface's palette
 * groups them, each with the smith command line it stands for under its title, and runs the
 * one the person picks. A command of the editor only says so.
 * @param contributed - the commands `package.json` contributes, as written there
 * @param views - the views of the MonoSmith container, in order, for the Go to section
 */
export async function showCommandPalette (contributed: ReadonlyArray<{ command: string, title: string }>, views: readonly PaletteView[]): Promise<void> {
  const commands: ContributedCommand[] = [
    ...contributed.map((command) => ({ command: command.command, title: titleOf(command.title) })),
    ...views.map((view) => ({ command: `${view.id}.focus`, title: say('vscode.command-palette.go-to', { view: view.name }) })),
  ]
  const goTo = new Set(views.map((view) => `${view.id}.focus`))
  type Item = vscode.QuickPickItem & { command?: string }
  const items: Item[] = toPaletteGroups(commands).flatMap((group): Item[] => [
    { label: say(SECTION_WORDS[group.section]), kind: vscode.QuickPickItemKind.Separator },
    ...group.entries.map((entry) => ({
      label:   entry.title,
      detail:  entry.line ?? say(goTo.has(entry.command) ? 'vscode.command-palette.a-view' : 'vscode.command-palette.in-the-editor'),
      command: entry.command,
    })),
  ])
  const picked = await vscode.window.showQuickPick(items, {
    title:         say('vscode.command-palette.title'),
    placeHolder:   say('vscode.command-palette.placeholder'),
    matchOnDetail: true,
  })
  if (picked?.command !== undefined) await vscode.commands.executeCommand(picked.command)
}
