import { describe, expect, it } from 'vitest'
import { toAdoptArguments, toMoveArguments } from './adopt-arguments.mapper.ts'

describe('toMoveArguments', () => {
  it('passes the folder so smith never asks, and commits nothing by default', () => {
    expect(toMoveArguments('apps/web', false)).toEqual(['adopt', '--init', '--move', 'apps/web'])
  })

  it('adds --commit only when it was ticked', () => {
    expect(toMoveArguments('none', true)).toEqual(['adopt', '--init', '--move', 'none', '--commit'])
  })
})

describe('toMoveArguments with the answers about the split', () => {
  it('passes the dependencies sent to the workspace and the ESLint answer, so smith does not ask again', () => {
    expect(toMoveArguments('apps/web', true, { hoist: ['nock', 'supertest'], eslint: 'shared' }))
      .toEqual(['adopt', '--init', '--move', 'apps/web', '--hoist', 'nock,supertest', '--eslint-setup', 'shared', '--commit'])
  })

  it('passes only what was asked', () => {
    expect(toMoveArguments('apps/web', false, { hoist: [], eslint: undefined })).toEqual(['adopt', '--init', '--move', 'apps/web'])
    expect(toMoveArguments('apps/web', false, { hoist: [], eslint: 'keep' })).toEqual(['adopt', '--init', '--move', 'apps/web', '--eslint-setup', 'keep'])
  })
})

describe('toAdoptArguments', () => {
  it('previews with a dry run that writes nothing', () => {
    expect(toAdoptArguments('preview')).toEqual(['adopt', '--dry-run'])
  })

  it('starts with smith alone, whose wizard asks before it writes', () => {
    expect(toAdoptArguments('start')).toEqual([])
  })
})
