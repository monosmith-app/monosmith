import { describe, expect, it } from 'vitest'
import { connectInProcess } from '../host/index.ts'
import type { SmithPlugin } from '../serve/index.ts'
import { checkPlugin } from './check-plugin.use-case.ts'
import { checkScheduling } from './check-scheduling.use-case.ts'
import { ConformanceReport } from './conformance-report.model.ts'

/** A language that declares a concurrency; with `stuck`, its discover takes seconds and ignores any abort. */
function scheduled (options: { concurrency?: number, cancellable?: boolean, stuck?: boolean } = {}): SmithPlugin {
  return {
    ...(options.cancellable === true) && { cancellable: true as const },
    describe: () => ({
      plugin:    { name: 'sched', version: '1.0.0' },
      kind:      'language',
      roles:     ['language'],
      manifests: ['sched.txt'],
      ...options.concurrency !== undefined && { concurrency: options.concurrency },
    }),
    discover: async () => {
      if (options.stuck === true) await new Promise(resolve => setTimeout(resolve, 3000))

      return { projects: [] }
    },
  }
}

const names = (report: ConformanceReport): string[] => report.results.map(result => result.check)

describe('the scheduling checks', () => {
  it('send a plugin that declares nothing three requests, one at a time, and do not check cancel', async () => {
    const client = await connectInProcess(scheduled(), { workspaceRoot: '/ws' })
    const report = await checkPlugin(client)
    expect(report.failures()).toBe('')
    expect(names(report)).toEqual(expect.arrayContaining(['declares a concurrency the protocol defines', 'answers 3 requests sent together, one at a time']))
    expect(names(report).some(name => name.includes('cancel'))).toBe(false)
    await client.close()
  })

  it('send a plugin that declares a concurrency that many requests together, and check its cancel', async () => {
    const client = await connectInProcess(scheduled({ concurrency: 4, cancellable: true }), { workspaceRoot: '/ws' })
    const report = await checkPlugin(client)
    expect(report.failures()).toBe('')
    expect(names(report)).toEqual(expect.arrayContaining([
      'accepts 4 requests at the same time',
      'answers a cancelled request with its result or the cancelled error',
      'ignores a cancel for a request it does not know',
    ]))
    await client.close()
  })

  it('report a concurrency the protocol does not define', async () => {
    const client = await connectInProcess(scheduled({ concurrency: -5 }), { workspaceRoot: '/ws' })
    const report = await checkPlugin(client)
    expect(report.failures()).toContain('declares a concurrency the protocol defines: concurrency -5 is below -1')
    await client.close()
  })

  it('report a plugin that advertises cancel and does not stop', async () => {
    const client = await connectInProcess(scheduled({ cancellable: true, stuck: true }), { workspaceRoot: '/ws', cancelGraceMs: 150 })
    const report = new ConformanceReport(client.info.plugin)
    await checkScheduling(client, report, 300)
    expect(report.failures()).toContain('answers a cancelled request with its result or the cancelled error')
  })
})
