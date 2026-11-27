import { describe, expect, it } from 'vitest'
import { pickArguments, repairArguments, updateArguments } from './update-arguments.policy.ts'

describe('updateArguments', () => {
  it('names one package, as smith deps up takes it', () => {
    expect(updateArguments('eslint')).toEqual(['deps', 'up', 'eslint'])
    expect(updateArguments('@types/node')).toEqual(['deps', 'up', '@types/node'])
    expect(updateArguments('golang.org/x/text')).toEqual(['deps', 'up', 'golang.org/x/text'])
  })

  it('names several marked packages in one command', () => {
    expect(updateArguments(['axios', 'zod'])).toEqual(['deps', 'up', 'axios', 'zod'])
  })

  it('adds --major for a major release', () => {
    expect(updateArguments('cobra', true)).toEqual(['deps', 'up', 'cobra', '--major'])
  })

  it('refuses what could be an option or is not a name', () => {
    expect(updateArguments('--yes')).toBeUndefined()
    expect(updateArguments('-y')).toBeUndefined()
    expect(updateArguments('a b')).toBeUndefined()
    expect(updateArguments('')).toBeUndefined()
    expect(updateArguments(['zod', '--yes'])).toBeUndefined()
    expect(updateArguments([])).toBeUndefined()
  })
})

describe('pickArguments', () => {
  it('lets the person pick in the terminal', () => {
    expect(pickArguments()).toEqual(['deps', 'up'])
  })
})

describe('repairArguments', () => {
  it('opens the repair, which asks what to repair', () => {
    expect(repairArguments()).toEqual(['deps', 'repair'])
  })
})
