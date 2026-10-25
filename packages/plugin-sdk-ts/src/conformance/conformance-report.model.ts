import type { Peer } from '../protocol/index.ts'

/** One check: its name, whether it held, and, when it did not, what is wrong. */
export interface CheckResult {
  check:   string
  ok:      boolean
  detail?: string
}

/**
 * Every check run against one plugin. `JSON.stringify(report)` gives the shape
 * `smith plugin check --json` prints: `{ plugin, results }`.
 */
export class ConformanceReport {
  readonly plugin:  Peer
  readonly results: CheckResult[] = []

  constructor (plugin: Peer) {
    this.plugin = plugin
  }

  /** Whether every check passed. */
  get ok (): boolean {
    return this.results.every(result => result.ok)
  }

  /** The failed checks, one per line: `check: detail`. */
  failures (): string {
    return this.results.filter(result => !result.ok).map(result => `${result.check}: ${result.detail ?? ''}`).join('\n')
  }

  /** Records a check. The detail is kept only for a check that failed. */
  add (check: string, ok: boolean, detail: string): void {
    this.results.push(ok ? { check, ok } : { check, ok, detail })
  }
}
