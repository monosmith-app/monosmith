import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync, existsSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { FileExistsError } from './file-exists.error.ts'
import { writeTemplates } from './write-templates.use-case.ts'

let root: string
beforeEach(() => { root = mkdtempSync(join(tmpdir(), 'sdk-templating-')) })
afterEach(() => { rmSync(root, { recursive: true, force: true }) })

// The first three tests are the Go SDK's templating tests.
describe('writeTemplates', () => {
  it('renders paths and bodies, and leaves GitHub syntax alone', async () => {
    const written = await writeTemplates(root, {
      '[[.Dir]]/main.go':    'package main // [[.Name]]',
      '[[.Dir]]/ci.yml':     'run: ${{ secrets.TOKEN }}',
      '[[.Dir]]/sub/doc.go': 'package sub',
    }, { Dir: 'apps/web', Name: 'web' })
    expect(written).toEqual(['apps/web/ci.yml', 'apps/web/main.go', 'apps/web/sub/doc.go'])
    expect(readFileSync(join(root, 'apps', 'web', 'main.go'), 'utf8')).toBe('package main // web')
    expect(readFileSync(join(root, 'apps', 'web', 'ci.yml'), 'utf8')).toBe('run: ${{ secrets.TOKEN }}')
  })

  it('refuses to overwrite a file and writes nothing', async () => {
    writeFileSync(join(root, 'b.txt'), 'mine')
    const refused = expect(writeTemplates(root, { 'a.txt': 'x', 'b.txt': 'y' })).rejects
    await refused.toBeInstanceOf(FileExistsError)
    await refused.toThrow('b.txt: file already exists')
    expect(existsSync(join(root, 'a.txt'))).toBe(false)
    expect(readFileSync(join(root, 'b.txt'), 'utf8')).toBe('mine')
  })

  it('fails on a missing key, writing nothing', async () => {
    await expect(writeTemplates(root, { ok: 'fine', a: '[[.Missing]]' }, {})).rejects.toThrow('map has no entry for key "Missing"')
    expect(existsSync(join(root, 'ok'))).toBe(false)
  })

  it('refuses a target outside the root', async () => {
    mkdirSync(join(root, 'inside'))
    await expect(writeTemplates(join(root, 'inside'), { '[[.Up]]/x': 'y' }, { Up: '..' })).rejects.toThrow('outside the project root')
    await expect(writeTemplates(root, { '../escape': 'y' })).rejects.toThrow('outside the project root')
  })
})
