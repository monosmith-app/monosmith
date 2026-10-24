/* eslint-disable regexp/no-super-linear-backtracking -- lines are capped at MAX_LINE below, so the backtracking is bounded */
import type { ParsedDiagnostic } from './parsed-diagnostic.contract.ts'

/** Reads the output of one task line by line; some formats need the lines before. */
export interface DiagnosticLineParser {
  /** Feeds the next line; returns the failure it completes, if it completes one. */
  feed: (line: string) => ParsedDiagnostic | undefined
}

// eslint-disable-next-line no-control-regex
const ANSI = /\u{1B}\[[0-9;]*[A-Z]/giu
/** Longer lines are minified bundles and the like, never a failure message. */
const MAX_LINE = 2000
const FILE = String.raw`(?:[a-z]:)?[^\s:()'"<>|]+\.[a-z0-9]+`
/** tsc: `src/a.ts(12,5): error TS2322: message` */
const TSC_PAREN = /^(?<file>\S.*?)\((?<line>\d+),(?<col>\d+)\):\s+(?<sev>error|warning|info)\s+(?<code>[A-Z]+\d+):\s*(?<msg>.+)$/
/** tsc with colors off but `--pretty`: `src/a.ts:12:5 - error TS2322: message` */
const TSC_PRETTY = /^(?<file>\S.*?):(?<line>\d+):(?<col>\d+)\s+-\s+(?<sev>error|warning|info)\s+(?<code>[A-Z]+\d+):\s*(?<msg>.+)$/
/** go, gcc, clang, ruff, mypy and the like: `path:line:col: message`, the column and a severity optional. */
const GENERIC = new RegExp(String.raw`^(?<file>${FILE}):(?<line>\d+)(?::(?<col>\d+))?:\s*(?:(?<sev>fatal error|error|warning|warn|info|note)\b\s*:?\s*)?(?<msg>\S.*)$`, 'i')
/** ESLint's stylish format: a path on a line of its own, then ` 12:5  error  message  rule`. */
const STYLISH_HEADER = new RegExp(`^(?<file>${FILE})$`, 'i')
const STYLISH_ITEM = /^\s+(?<line>\d+):(?<col>\d+)\s+(?<sev>error|warning)\s+(?<msg>.+?)(?:\s{2,}(?<code>[\w@/.-]+))?\s*$/
/** rustc: `error[E0425]: message`, then ` --> src/main.rs:3:5`. */
const RUST_HEAD = /^\s*(?<sev>error|warning)(?:\[(?<code>\w+)\])?:\s+(?<msg>.+)$/
const RUST_WHERE = /^\s*-->\s+(?<file>.+?):(?<line>\d+):(?<col>\d+)\s*$/

/**
 * Creates a parser for the output of one task. It knows the formats of tsc
 * (both styles), go and gcc style `path:line:col: message`, ESLint's stylish
 * report and rustc. A line it does not know is not a failure.
 */
export function createDiagnosticLineParser (): DiagnosticLineParser {
  let stylishFile: string | undefined
  let rust: { severity: ParsedDiagnostic['severity'], message: string, code?: string } | undefined

  return {
    feed: (raw) => {
      if (raw.length > MAX_LINE) return
      const line = raw.replaceAll(ANSI, '').replace(/\s+$/, '')
      if (line.trim() === '') {
        stylishFile = undefined
        rust = undefined

        return
      }
      const fromRust = readRust(line, rust)
      rust = fromRust.pending
      if (fromRust.found !== undefined) return fromRust.found
      const stylish = STYLISH_ITEM.exec(line)?.groups
      if (stylish !== undefined && stylishFile !== undefined) {
        return make(stylishFile, stylish.line, stylish.col, stylish.sev, stylish.msg, stylish.code)
      }
      const header = STYLISH_HEADER.exec(line)?.groups
      if (header !== undefined) {
        stylishFile = header.file

        return
      }
      stylishFile = undefined
      const grouped = (TSC_PAREN.exec(line) ?? TSC_PRETTY.exec(line) ?? GENERIC.exec(line))?.groups

      return grouped === undefined ? undefined : make(grouped.file, grouped.line, grouped.col, grouped.sev, grouped.msg, grouped.code)
    },
  }
}

function readRust (line: string, pending: { severity: ParsedDiagnostic['severity'], message: string, code?: string } | undefined): { found?: ParsedDiagnostic, pending: typeof pending } {
  const where = RUST_WHERE.exec(line)?.groups
  if (where !== undefined && pending !== undefined) {
    return { found: { file: where.file ?? '', line: Number(where.line), column: Number(where.col), ...pending }, pending: undefined }
  }
  const head = RUST_HEAD.exec(line)?.groups
  if (head !== undefined) {
    return { pending: { severity: severityOf(head.sev), message: head.msg ?? '', ...(head.code !== undefined && { code: head.code }) } }
  }

  return { pending: where === undefined ? pending : undefined }
}

function make (file: string | undefined, line: string | undefined, column: string | undefined, severity: string | undefined, message: string | undefined, code: string | undefined): ParsedDiagnostic {
  return {
    file:     file ?? '',
    line:     Number(line),
    column:   column === undefined ? 1 : Number(column),
    severity: severityOf(severity),
    message:  (message ?? '').trim(),
    ...(code !== undefined && code !== '' && { code }),
  }
}

function severityOf (word: string | undefined): ParsedDiagnostic['severity'] {
  const lower = word?.toLowerCase()
  if (lower === 'warning' || lower === 'warn') return 'warning'

  return lower === 'info' || lower === 'note' ? 'info' : 'error'
}
