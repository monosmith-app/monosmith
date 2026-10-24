import { describe, expect, it } from 'vitest'
import { createDiagnosticLineParser } from './parse-diagnostic.algorithm.ts'

function parseAll (lines: string[]) {
  const parser = createDiagnosticLineParser()

  return lines.flatMap((line) => parser.feed(line) ?? [])
}

describe('createDiagnosticLineParser', () => {
  it('reads a tsc failure', () => {
    expect(parseAll(["src/app/main.ts(12,5): error TS2322: Type 'string' is not assignable to type 'number'."])).toEqual([
      { file: 'src/app/main.ts', line: 12, column: 5, severity: 'error', code: 'TS2322', message: "Type 'string' is not assignable to type 'number'." },
    ])
  })

  it('reads a tsc failure in the pretty style, through the colors', () => {
    expect(parseAll(['\u{1B}[96msrc/a.ts\u{1B}[0m:\u{1B}[93m3\u{1B}[0m:\u{1B}[93m7\u{1B}[0m - \u{1B}[91merror\u{1B}[0m\u{1B}[90m TS2304: \u{1B}[0mCannot find name x.'])).toMatchObject([
      { file: 'src/a.ts', line: 3, column: 7, severity: 'error', code: 'TS2304', message: 'Cannot find name x.' },
    ])
  })

  it('reads go and gcc style lines, with and without a column or a severity', () => {
    expect(parseAll([
      './main.go:12:5: undefined: x',
      'src/a.c:3:9: warning: unused variable',
      'tool.py:40: SyntaxWarning: invalid escape',
    ])).toMatchObject([
      { file: './main.go', line: 12, column: 5, severity: 'error', message: 'undefined: x' },
      { file: 'src/a.c', line: 3, column: 9, severity: 'warning', message: 'unused variable' },
      { file: 'tool.py', line: 40, column: 1, severity: 'error', message: 'SyntaxWarning: invalid escape' },
    ])
  })

  it('reads a Windows path with a drive letter', () => {
    expect(parseAll([String.raw`C:\repo\apps\api\main.go:7:2: missing return`])).toMatchObject([
      { file: String.raw`C:\repo\apps\api\main.go`, line: 7, column: 2 },
    ])
  })

  it('reads the ESLint stylish report, a file at a time', () => {
    const found = parseAll([
      '/repo/apps/web/src/a.ts',
      '  3:10  error  Unexpected any  @typescript-eslint/no-explicit-any',
      "  8:1   warning  'x' is assigned a value but never used  no-unused-vars",
      '',
      '/repo/apps/web/src/b.ts',
      '  1:1  error  Parsing error: unexpected token',
      '',
      '✖ 3 problems (2 errors, 1 warning)',
    ])
    expect(found).toMatchObject([
      { file: '/repo/apps/web/src/a.ts', line: 3, column: 10, severity: 'error', message: 'Unexpected any', code: '@typescript-eslint/no-explicit-any' },
      { file: '/repo/apps/web/src/a.ts', line: 8, column: 1, severity: 'warning', code: 'no-unused-vars' },
      { file: '/repo/apps/web/src/b.ts', line: 1, column: 1, severity: 'error', message: 'Parsing error: unexpected token' },
    ])
  })

  it('does not read stylish items that belong to no file', () => {
    expect(parseAll(['  3:10  error  Unexpected any  rule'])).toEqual([])
  })

  it('reads rustc, which names the place on the line after the message', () => {
    expect(parseAll([
      'error[E0425]: cannot find value `y` in this scope',
      ' --> src/main.rs:3:5',
      '  |',
    ])).toEqual([
      { file: 'src/main.rs', line: 3, column: 5, severity: 'error', code: 'E0425', message: 'cannot find value `y` in this scope' },
    ])
  })

  it('leaves ordinary output alone', () => {
    expect(parseAll([
      'compiling...',
      '12:30:45 starting the server',
      'Error: something broke',
      'listening on http://localhost:3000',
      'PASS src/a.test.ts',
      '',
    ])).toEqual([])
  })
})
