import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

interface Pattern { regexp: string, file?: number, line?: number, column?: number, severity?: number, code?: number, message?: number, loop?: boolean }
interface Matcher { name: string, pattern: Pattern | Pattern[] }
const manifest = JSON.parse(readFileSync(new URL('../../package.json', import.meta.url), 'utf8')) as {
  contributes: { problemMatchers: Matcher[], taskDefinitions: Array<{ type: string, required: string[] }> }
}

function matcher (name: string): Pattern[] {
  const found = manifest.contributes.problemMatchers.find((entry) => entry.name === name)
  if (found === undefined) throw new Error(`no problem matcher ${name}`)

  return Array.isArray(found.pattern) ? found.pattern : [found.pattern]
}

/** What VS Code would take from a line with the single-line pattern of a matcher. */
function read (name: string, line: string): Record<string, string | undefined> | undefined {
  const [pattern] = matcher(name)
  const match = pattern === undefined ? null : new RegExp(pattern.regexp).exec(line)
  if (match === null || pattern === undefined) return undefined
  const at = (index: number | undefined): string | undefined => (index === undefined ? undefined : match[index])

  return { file: at(pattern.file), line: at(pattern.line), column: at(pattern.column), severity: at(pattern.severity), code: at(pattern.code), message: at(pattern.message) }
}

describe('the problem matchers of the manifest', () => {
  it('declare the smith task type with a required task', () => {
    expect(manifest.contributes.taskDefinitions).toMatchObject([{ type: 'smith', required: ['task'] }])
  })

  describe('$smith', () => {
    it('reads go, with the prefix smith puts before the lines of a task', () => {
      expect(read('smith', 'api:build │ ./main.go:12:5: undefined: x')).toMatchObject({ file: './main.go', line: '12', column: '5', message: 'undefined: x' })
    })

    it('reads a severity and a line without a column', () => {
      expect(read('smith', 'src/a.c:3:9: warning: unused variable')).toMatchObject({ file: 'src/a.c', line: '3', column: '9', severity: 'warning', message: 'unused variable' })
      expect(read('smith', 'tool.py:40: bad escape')).toMatchObject({ file: 'tool.py', line: '40', column: undefined })
    })

    it('leaves ordinary output alone', () => {
      expect(read('smith', 'compiling...')).toBeUndefined()
      expect(read('smith', 'web:build ▶ npm run build --silent')).toBeUndefined()
      expect(read('smith', 'listening on http://localhost:3000')).toBeUndefined()
    })
  })

  describe('$smith-tsc', () => {
    it('reads a tsc failure with its code', () => {
      expect(read('smith-tsc', "web:typecheck │ src/main.ts(12,5): error TS2322: Type 'string' is not assignable.")).toMatchObject({
        file: 'src/main.ts', line: '12', column: '5', severity: 'error', code: 'TS2322', message: "Type 'string' is not assignable.",
      })
    })
  })

  describe('$smith-eslint', () => {
    it('reads the file line and then the items of a stylish report', () => {
      const [header, item] = matcher('smith-eslint')
      expect(new RegExp(header?.regexp ?? '').exec('web:lint │ /repo/apps/web/src/a.ts')?.[1]).toBe('/repo/apps/web/src/a.ts')
      const found = new RegExp(item?.regexp ?? '').exec('web:lint │   3:10  error  Unexpected any  @typescript-eslint/no-explicit-any')
      expect(found?.slice(1, 6)).toEqual(['3', '10', 'error', 'Unexpected any', '@typescript-eslint/no-explicit-any'])
      expect(item?.loop).toBe(true)
    })
  })
})
