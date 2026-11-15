import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { resolveSourcePath } from './resolve-source-path.policy.ts'

const root = join(process.cwd(), 'repo')
const project = join(root, 'apps', 'web')
const existing = new Set([join(project, 'src', 'a.ts'), join(root, 'tsconfig.json')])
const exists = (path: string): boolean => existing.has(path)

describe('resolveSourcePath', () => {
  it('tries the project directory before the workspace folder', () => {
    expect(resolveSourcePath({ file: 'src/a.ts', bases: [project, root], exists })).toBe(join(project, 'src', 'a.ts'))
    expect(resolveSourcePath({ file: 'tsconfig.json', bases: [project, root], exists })).toBe(join(root, 'tsconfig.json'))
  })

  it('understands a leading ./', () => {
    expect(resolveSourcePath({ file: './src/a.ts', bases: [project], exists })).toBe(join(project, 'src', 'a.ts'))
  })

  it('keeps an absolute path that exists', () => {
    const absolute = join(project, 'src', 'a.ts')
    expect(resolveSourcePath({ file: absolute, bases: [], exists })).toBe(absolute)
  })

  it('is undefined for a file that is nowhere', () => {
    expect(resolveSourcePath({ file: 'src/missing.ts', bases: [project, root], exists })).toBeUndefined()
    expect(resolveSourcePath({ file: join(root, 'gone.ts'), bases: [], exists })).toBeUndefined()
  })
})
