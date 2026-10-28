import { describe, expect, it } from 'vitest'
import { toAdoptArguments } from './adopt-arguments.mapper.ts'

describe('toAdoptArguments', () => {
  it('previews with a dry run that writes nothing', () => {
    expect(toAdoptArguments('preview')).toEqual(['adopt', '--dry-run'])
  })

  it('starts with smith alone, whose wizard asks before it writes', () => {
    expect(toAdoptArguments('start')).toEqual([])
  })
})
