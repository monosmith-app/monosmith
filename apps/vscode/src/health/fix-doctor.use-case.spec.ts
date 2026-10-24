import { describe, expect, it } from 'vitest'
import type { SmithConnection } from '../smith-client/index.ts'
import { fixDoctor } from './fix-doctor.use-case.ts'

const PROPOSED = 'smith doctor --fix exited with code 1 and printed no JSON: x\n\n● ignore the cache directory\n+.smith/\n\nPass --yes to apply these fixes.'

function fakeSmith (fixes: 'proposed' | 'none', applied?: { ok: boolean }) {
  const calls: Array<{ method: string, params: unknown }> = []
  const listeners = new Set<(event: unknown) => void>()
  const connection: SmithConnection = {
    request: async <T>(method: string, params?: unknown): Promise<T> => {
      calls.push({ method, params })
      if (method === 'doctor.fixes') throw new Error(fixes === 'proposed' ? PROPOSED : 'Nothing doctor can fix itself')
      const run = (params as { run: string }).run
      for (const listener of listeners) listener({ event: 'task.output', run, line: 'wrote .gitignore' })

      return (applied ?? { ok: true }) as T
    },
    onEvent: (handler) => {
      listeners.add(handler)

      return () => { listeners.delete(handler) }
    },
    close: () => {},
  }

  return { connection, calls }
}

describe('fixDoctor', () => {
  it('asks before it writes anything, and writes only on a yes', async () => {
    const { connection, calls } = fakeSmith('proposed')
    const seen: string[][] = []
    const lines: string[] = []
    const outcome = await fixDoctor(connection, async (preview) => {
      seen.push(preview.titles)

      return true
    }, (line) => { lines.push(line) })
    expect(outcome).toBe('fixed')
    expect(seen).toEqual([['ignore the cache directory']])
    expect(calls.map((call) => call.method)).toEqual(['doctor.fixes', 'doctor.fix'])
    expect(lines).toEqual(['wrote .gitignore'])
  })

  it('does not call the fixing when the person says no', async () => {
    const { connection, calls } = fakeSmith('proposed')
    expect(await fixDoctor(connection, async () => false)).toBe('declined')
    expect(calls.map((call) => call.method)).toEqual(['doctor.fixes'])
  })

  it('does nothing when doctor proposes no fix', async () => {
    const { connection, calls } = fakeSmith('none')
    expect(await fixDoctor(connection, async () => true)).toBe('nothing')
    expect(calls).toHaveLength(1)
  })

  it('reports that a check without a fix still fails after the fixing', async () => {
    const { connection } = fakeSmith('proposed', { ok: false })
    expect(await fixDoctor(connection, async () => true)).toBe('incomplete')
  })
})
