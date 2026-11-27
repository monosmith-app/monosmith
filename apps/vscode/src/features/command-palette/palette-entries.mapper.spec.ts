import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { COMMAND_LINES, PALETTE_SECTIONS } from './command-lines.config.ts'
import { commandLineOf, toPaletteGroups } from './palette-entries.mapper.ts'

const PACKAGE_JSON = fileURLToPath(new URL('../../../package.json', import.meta.url))
const manifest = JSON.parse(readFileSync(PACKAGE_JSON, 'utf8')) as { contributes: { commands: Array<{ command: string, title: string }> } }
const contributed = manifest.contributes.commands.map((command) => command.command.replace(/^monosmith\./, ''))

describe('the command lines', () => {
  it('say what every command of package.json stands for', () => {
    expect(contributed.filter((id) => COMMAND_LINES[id] === undefined)).toEqual([])
  })

  it('name no command that package.json lacks', () => {
    expect(Object.keys(COMMAND_LINES).filter((id) => !contributed.includes(id))).toEqual([])
  })

  it('are smith command lines, or nothing for a command of the editor only', () => {
    const lines = Object.values(COMMAND_LINES).flatMap((entry) => entry.line ?? [])
    expect(lines.filter((line) => !line.startsWith('smith') && !line.startsWith('npm '))).toEqual([])
  })
})

describe('commandLineOf', () => {
  it('reads a full id and a short one', () => {
    expect(commandLineOf('monosmith.buildAll')).toBe('smith run build')
    expect(commandLineOf('release.dryRun')).toBe('smith release --dry-run')
  })

  it('has none for a command of the editor only, or an unknown one', () => {
    expect(commandLineOf('monosmith.showOutput')).toBeUndefined()
    expect(commandLineOf('monosmith.nothing')).toBeUndefined()
  })
})

describe('toPaletteGroups', () => {
  const lines = {
    buildAll: { section: 'build', line: 'smith run build' },
    lintAll:  { section: 'build', line: 'smith run lint' },
    doctor:   { section: 'health', line: 'smith doctor' },
    new:      { section: 'start', line: 'smith new --here' },
  } as const

  it('groups the commands in the order of the palette, by title, with their lines', () => {
    const groups = toPaletteGroups([
      { command: 'monosmith.doctor', title: 'Run Doctor' },
      { command: 'monosmith.lintAll', title: 'Lint all' },
      { command: 'monosmith.buildAll', title: 'Build all' },
      { command: 'monosmith.new', title: 'New workspace' },
    ], lines)
    expect(groups.map((group) => group.section)).toEqual(['start', 'build', 'health'])
    expect(groups[1]?.entries).toEqual([
      { command: 'monosmith.buildAll', title: 'Build all', section: 'build', line: 'smith run build' },
      { command: 'monosmith.lintAll', title: 'Lint all', section: 'build', line: 'smith run lint' },
    ])
  })

  it('puts a command the table does not know under Help, without a line', () => {
    const groups = toPaletteGroups([{ command: 'monosmith.home.focus', title: 'Go to Home' }], lines)
    expect(groups).toEqual([{ section: 'help', entries: [{ command: 'monosmith.home.focus', title: 'Go to Home', section: 'help' }] }])
  })

  it('keeps the eight sections of the terminal palette', () => {
    expect(PALETTE_SECTIONS).toEqual(['start', 'build', 'projects', 'dependencies', 'release', 'health', 'goto', 'help'])
  })
})
