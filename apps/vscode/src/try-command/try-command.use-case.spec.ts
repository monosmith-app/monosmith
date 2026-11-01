import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { tryCommand } from './try-command.use-case.ts'

const PASSED: SmithRun = {
  code:   0,
  stdout: JSON.stringify({ command: ['add', 'package', 'billing'], commandOk: true, changed: ['A  packages/billing/package.json'], checks: [{ name: 'doctor', ok: true }] }),
  stderr: '',
}
const FAILED: SmithRun = {
  code:   1,
  stdout: JSON.stringify({ command: ['add', 'package', 'billing'], commandOk: true, changed: [], checks: [{ name: 'doctor', ok: false, tail: 'two projects share a name' }] }),
  stderr: '',
}

function fake (runs: SmithRun[], words: string[] | undefined, apply: boolean) {
  const calls: string[][] = []
  const pages: string[] = []

  return {
    calls,
    pages,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? { code: 0, stdout: '', stderr: '' }
      },
      ask:   async () => words,
      show:  async (page: string) => { pages.push(page) },
      offer: async () => apply,
    },
  }
}

describe('tryCommand', () => {
  it('tries in a copy, shows the page and runs it for real only on a yes', async () => {
    const f = fake([PASSED], ['add', 'package', 'billing'], true)
    expect(await tryCommand(f.deps)).toEqual({ kind: 'tried', passed: true, applied: true })
    expect(f.calls).toEqual([['try', 'add', 'package', 'billing', '--json'], ['try', 'add', 'package', 'billing', '--yes']])
    expect(f.pages[0]).toContain('# Try: smith add package billing')
    expect(f.pages[0]).toContain('`A  packages/billing/package.json`')
  })

  it('leaves the files alone when the person says no', async () => {
    const f = fake([PASSED], ['add', 'package', 'billing'], false)
    expect(await tryCommand(f.deps)).toEqual({ kind: 'tried', passed: true, applied: false })
    expect(f.calls).toHaveLength(1)
  })

  it('shows a failed check and offers nothing', async () => {
    const f = fake([FAILED], ['add', 'package', 'billing'], true)
    expect(await tryCommand(f.deps)).toEqual({ kind: 'tried', passed: false, applied: false })
    expect(f.pages[0]).toContain('two projects share a name')
    expect(f.calls).toHaveLength(1)
  })

  it('returns smith\'s words when it printed no result', async () => {
    const f = fake([{ code: 1, stdout: '', stderr: '✖ nothing to try\n  → name a command\n' }], ['wat'], true)
    expect(await tryCommand(f.deps)).toEqual({ kind: 'refused', message: 'nothing to try Try: name a command' })
  })

  it('does nothing when the person cancels', async () => {
    const f = fake([], undefined, true)
    expect(await tryCommand(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.calls).toEqual([])
  })
})
