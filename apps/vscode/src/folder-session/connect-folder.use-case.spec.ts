import { describe, expect, it } from 'vitest'
import type { SmithLookup, SmithSession } from '../smith-client/index.ts'
import { connectFolder } from './connect-folder.use-case.ts'

const session = { version: '1.56.0' } as unknown as SmithSession
const found: SmithLookup = { location: { executable: '/usr/bin/smith', source: 'path' }, pathSmith: '/usr/bin/smith', skippedUntrusted: false }

describe('connectFolder', () => {
  it('is ready with the session when smith starts', async () => {
    const outcome = await connectFolder({ locate: () => found, start: async () => session }, '/repo')
    expect(outcome.phase).toBe('ready')
    expect(outcome.session).toBe(session)
    expect(outcome.location?.source).toBe('path')
  })

  it('is missing when there is no smith', async () => {
    const outcome = await connectFolder({ locate: () => ({ skippedUntrusted: false }), start: async () => session }, '/repo')
    expect(outcome.phase).toBe('missing')
  })

  it('is untrusted when the only smith is the folder\'s own and the folder is not trusted', async () => {
    const outcome = await connectFolder({ locate: () => ({ skippedUntrusted: true }), start: async () => session }, '/repo')
    expect(outcome.phase).toBe('untrusted')
  })

  it('is failed, with the reason, when smith cannot start', async () => {
    const outcome = await connectFolder({ locate: () => found, start: async () => { throw new Error('spawn EACCES') } }, '/repo')
    expect(outcome.phase).toBe('failed')
    expect(outcome.error).toBe('spawn EACCES')
    expect(outcome.location?.executable).toBe('/usr/bin/smith')
  })

  it('names the setting and its path when the monosmith.path smith cannot start', async () => {
    const outcome = await connectFolder({ locate: () => ({ location: { executable: '/nope/smith', source: 'setting' }, skippedUntrusted: false }), start: async () => { throw new Error('spawn ENOENT') } }, '/repo')
    expect(outcome.error).toBe('the monosmith.path setting (/nope/smith) did not start: spawn ENOENT')
  })

  it('does not start anything when there is no smith', async () => {
    let started = false
    await connectFolder({
      locate: () => ({ skippedUntrusted: false }),
      start:  async () => {
        started = true

        return session
      },
    }, '/repo')
    expect(started).toBe(false)
  })
})
