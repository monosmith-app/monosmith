import { describe, expect, it } from 'vitest'
import { protocolSchema } from '../protocol/index.ts'
import { capabilitiesOf } from './capabilities-of.algorithm.ts'
import { PLUGIN_MEMBERS } from './plugin-members.config.ts'
import type { SmithPlugin } from './plugin.contract.ts'

describe('PLUGIN_MEMBERS', () => {
  it('covers every method of the schema except the two the serve loop answers itself and the cancel notification it handles', () => {
    const methods = Object.keys(protocolSchema['x-methods']).filter(m => m !== 'initialize' && m !== 'shutdown' && m !== 'cancel')
    expect(new Set(Object.keys(PLUGIN_MEMBERS))).toEqual(new Set(methods))
  })

  it('derives every capability of the schema from some methods, and no other', () => {
    const everything: SmithPlugin = {
      describe:     () => ({ plugin: { name: 'a', version: 'b' }, kind: 'language' }),
      discover:     () => ({ projects: [] }),
      generate:     () => ({ files: [] }),
      writeVersion: () => ({ files: [] }),
      publish:      () => ({ published: false }),
      published:    () => ({}),
      ci:           () => ({}),
      doctor:       () => ({ findings: [] }),
      contribute:   () => ({}),
      remove:       () => ({}),
      changeKind:   () => ({}),
      install:      () => ({}),
      uninstall:    () => ({}),
      declared:     () => ({ declarations: [] }),
      rewrite:      () => ({}),
      latest:       () => ({}),
      cancellable:  true,
    }
    const rows = Object.values(protocolSchema['x-methods'])
    const wanted = new Set(rows.map(row => row.capability).filter(c => c !== null))
    const advertised = new Set(Object.keys(capabilitiesOf(everything)))
    expect(advertised).toEqual(wanted)
    expect(wanted.size).toBe(14)
  })
})
