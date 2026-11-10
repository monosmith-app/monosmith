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

describe('toAdoptArguments', () => {
  it('previews with a dry run that writes nothing', () => {
    expect(toAdoptArguments('preview')).toEqual(['adopt', '--dry-run'])
  })

  it('starts with smith alone, whose wizard asks before it writes', () => {
    expect(toAdoptArguments('start')).toEqual([])
  })
})
