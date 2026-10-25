import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import type { GenerateParams } from '../protocol/index.ts'
import { scaffoldData } from './scaffold-data.mapper.ts'
import { scaffoldProject } from './scaffold-project.use-case.ts'

let root: string
beforeEach(() => { root = mkdtempSync(join(tmpdir(), 'sdk-scaffold-')) })
afterEach(() => { rmSync(root, { recursive: true, force: true }) })

const params: GenerateParams = { kind: 'lib', name: 'core', directory: 'libs/core', workspace: { name: 'acme', scope: '@acme' }, stack: 'widget-plain' }

describe('scaffoldData', () => {
  it('names the fields a template can use', () => {
    expect(scaffoldData(params)).toEqual({ Dir: 'libs/core', Name: 'core', Kind: 'lib', Stack: 'widget-plain', Workspace: 'acme', Scope: '@acme', ModulePrefix: '' })
    expect(scaffoldData({ ...params, stack: undefined }).Stack).toBe('')
  })
})

describe('scaffoldProject', () => {
  it('writes the templates into the project and reports the files', async () => {
    const result = await scaffoldProject({
      workspaceRoot: root,
      params,
      files:         { '[[.Dir]]/widget.json': '{"name": "[[.Name]]"[[if .Stack]], "stack": "[[.Stack]]"[[end]], "port": [[.Port]]}\n', '[[.Dir]]/README.md': '# [[.Name]]\n' },
      data:          { Port: 9000 },
      notes:         ['edit widget.json'],
    })
    expect(result).toEqual({ files: ['libs/core/README.md', 'libs/core/widget.json'], notes: ['edit widget.json'] })
    expect(readFileSync(join(root, 'libs', 'core', 'widget.json'), 'utf8')).toBe('{"name": "core", "stack": "widget-plain", "port": 9000}\n')
  })

  it('never overwrites and omits notes when there are none', async () => {
    await scaffoldProject({ workspaceRoot: root, params, files: { 'a.txt': 'one' } })
    await expect(scaffoldProject({ workspaceRoot: root, params, files: { 'b.txt': 'two', 'a.txt': 'three' } })).rejects.toThrow('a.txt: file already exists')
    expect(existsSync(join(root, 'b.txt'))).toBe(false)
    expect(await scaffoldProject({ workspaceRoot: root, params, files: { 'c.txt': '' } })).toEqual({ files: ['c.txt'] })
  })
})
