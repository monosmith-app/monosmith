import { describe, expect, it } from 'vitest'
import type { ComposeItem } from './compose-stack.contract.ts'
import { composeArguments, describePlan, parseReport, pickLabels, settleRivals, stackSummary, tickedTitles, toolsToInstall } from './compose-report.mapper.ts'

const catalogue = JSON.stringify({
  layers: [{ id: 'shell', title: 'Desktop shell', about: 'a program you install' }],
  items:  [
    { id: 'node/electron', kind: 'choice', name: 'electron', title: 'Electron', description: 'a desktop app', aliases: ['desktop'], layer: 'shell', language: 'node' },
    { id: '', kind: 'choice', name: 'broken' },
  ],
  check: { language: 'node', complete: false, conflict: 'react and electron are both on the framework axis: keep one' },
})

describe('parseReport', () => {
  it('reads the layers, the items with a name and the check', () => {
    const report = parseReport(catalogue)
    expect(report.layers[0]?.title).toBe('Desktop shell')
    expect(report.items.map((item) => item.name)).toEqual(['electron'])
    expect(report.check?.conflict).toContain('keep one')
    expect(report.plan).toBeUndefined()
  })

  it('reads a question with its suggested answer', () => {
    const report = parseReport(JSON.stringify({ items: [], check: { language: 'node', complete: false, question: { choice: 'azure-functions', axis: 'trigger', options: [{ name: 'http-trigger', suggested: true }] } } }))
    expect(report.check?.question?.options[0]).toMatchObject({ name: 'http-trigger', title: 'http-trigger', suggested: true })
  })

  it('refuses text that is not a catalogue', () => {
    expect(() => parseReport('[]')).toThrow(TypeError)
    expect(() => parseReport('nope')).toThrow(SyntaxError)
  })
})

describe('composeArguments', () => {
  it('asks for the catalogue alone, the check, or the plan of a named project', () => {
    expect(composeArguments({ picks: [] })).toEqual(['compose', '--json'])
    expect(composeArguments({ language: 'node', picks: ['react', 'eslint'] })).toEqual(['compose', '--json', '--lang', 'node', '--with', 'react,eslint'])
    expect(composeArguments({ language: 'node', picks: ['electron'], kind: 'app', name: 'desk' })).toEqual(['compose', '--json', '--lang', 'node', '--with', 'electron', '--kind', 'app', '--name', 'desk'])
  })
})

describe('pickLabels', () => {
  it('puts the aliases in the detail, so typing desktop finds Electron', () => {
    const [item] = parseReport(catalogue).items
    expect(pickLabels(item)).toMatchObject({ label: 'Electron', description: 'node · shell' })
    expect(pickLabels(item).detail).toContain('desktop')
  })

  it('says why an item cannot be used and which command turns a plugin on', () => {
    const labels = pickLabels({ id: 'plugin/slices', kind: 'plugin', name: 'slices', title: 'slices', description: '', aliases: [], layer: 'architecture', axis: '', many: false, language: '', suggested: false, isDefault: false, problem: '', route: 'smith plugin enable slices' })
    expect(labels.detail).toContain('smith plugin enable slices')
  })

  it('says which question an item answers when its layer has another name, so the label and the conflict agree', () => {
    const electron = pickLabels({ id: 'node/electron', kind: 'choice', name: 'electron', title: 'Electron', description: '', aliases: [], layer: 'shell', axis: 'framework', many: false, language: 'node', suggested: false, isDefault: false, problem: '', route: '' })
    expect(electron.description).toBe('node · shell (the framework choice)')
    const react = pickLabels({ id: 'node/react', kind: 'choice', name: 'react', title: 'React', description: '', aliases: [], layer: 'framework', axis: 'framework', many: false, language: 'node', suggested: false, isDefault: false, problem: '', route: '' })
    expect(react.description).toBe('node · framework')
  })
})

describe('describePlan', () => {
  it('lists the command, the files, the tools and the notes', () => {
    const text = describePlan({
      command: 'smith add app desk --lang node --with electron,eslint,vitest',
      argv:    [],
      files:   [{ action: 'create', path: 'apps/desk/', note: '' }],
      tools:   [
        { choice: 'node', name: 'node', title: 'Node.js', binary: 'node', state: 'installed', command: 'smith tools install node' },
        { choice: 'tauri', name: 'tauri-linux-libs', title: 'the Tauri system libraries', binary: 'pkg-config', state: 'missing', command: 'smith tools install tauri-linux-libs' },
      ],
      notes:      ['eslint and vitest are added by default'],
      recipeArgv: [],
      recipeNote: '',
    })
    expect(text).toContain('Run: smith add app desk')
    expect(text).toContain('create apps/desk/')
    expect(text).toContain('Node.js (installed, for node)')
    expect(text).toContain('the Tauri system libraries (missing, for tauri): smith tools install tauri-linux-libs')
    expect(text).toContain('added by default')
  })
})

describe('parseReport plan tools', () => {
  it('reads the title, the state and the install command that smith prints', () => {
    const report = parseReport(JSON.stringify({ items: [], plan: { command: 'smith add app a', argv: ['smith', 'add'], tools: [{ choice: 'tauri', name: 'tauri-linux-libs', title: 'the Tauri system libraries', binary: 'pkg-config', state: 'missing', command: 'smith tools install tauri-linux-libs' }] } }))
    expect(report.plan?.tools[0]).toMatchObject({ title: 'the Tauri system libraries', state: 'missing', command: 'smith tools install tauri-linux-libs' })
    expect(report.plan !== undefined && toolsToInstall(report.plan).map((tool) => tool.name)).toEqual(['tauri-linux-libs'])
  })
})

const row = (id: string, over: Partial<ComposeItem> = {}): ComposeItem => ({ id, kind: 'choice', name: id.split('/', 2)[1] ?? id, title: id, description: '', aliases: [], layer: 'framework', axis: 'framework', many: false, language: 'node', suggested: false, isDefault: false, problem: '', route: '', ...over })
const catalogueItems = [
  row('node/react', { title: 'React' }),
  row('node/electron', { title: 'Electron', layer: 'shell' }),
  row('node/vitest', { title: 'Vitest', layer: 'test', axis: 'test' }),
  row('node/eslint', { title: 'ESLint', layer: 'lint', axis: 'lint', many: true }),
  row('node/prettier', { title: 'Prettier', layer: 'lint', axis: 'lint', many: true }),
  row('node', { kind: 'language', name: 'node', title: 'TypeScript / JavaScript', axis: '', language: '', layer: 'language' }),
  row('python', { kind: 'language', name: 'python', title: 'Python', axis: '', language: '', layer: 'language' }),
]

describe('tickedTitles and stackSummary', () => {
  it('names what is ticked, so a search that hides the rows hides nothing', () => {
    expect(tickedTitles(catalogueItems, ['node/electron', 'node/react'])).toBe('React, Electron')
    expect(tickedTitles(catalogueItems, [])).toBe('')
  })

  it('names the stack with its language', () => {
    expect(stackSummary(catalogueItems, 'node', ['react', 'vitest'])).toBe('React + Vitest (TypeScript / JavaScript)')
  })
})

const item = (name: string, axis: string, many = false): ComposeItem => ({ id: `node/${name}`, kind: 'choice', name, title: name, description: '', aliases: [], layer: axis, axis, many, language: 'node', suggested: false, isDefault: false, problem: '', route: '' })

describe('settleRivals', () => {
  const items = [item('react', 'framework'), item('electron', 'framework'), item('tauri', 'shell'), item('vitest', 'test', true), item('jest', 'test', true)]

  it('lets the new pick take the place of the one before on a single-choice axis', () => {
    const settled = settleRivals(items, ['node/react', 'node/tauri'], ['node/react', 'node/tauri', 'node/electron'])
    expect(settled.ids).toEqual(['node/tauri', 'node/electron'])
    expect(settled.note).toBe('electron replaces react: a project has one framework.')
  })

  it('keeps picks on other axes and on an axis that takes several', () => {
    const settled = settleRivals(items, ['node/vitest'], ['node/vitest', 'node/jest', 'node/tauri'])
    expect(settled.ids).toEqual(['node/vitest', 'node/jest', 'node/tauri'])
    expect(settled.note).toBe('')
  })

  it('lets one language be ticked at a time', () => {
    expect(settleRivals(catalogueItems, ['node'], ['node', 'python']).ids).toEqual(['python'])
  })

  it('keeps the later one when two rivals are ticked in the same round', () => {
    expect(settleRivals(items, [], ['node/react', 'node/electron']).ids).toEqual(['node/electron'])
  })
})
