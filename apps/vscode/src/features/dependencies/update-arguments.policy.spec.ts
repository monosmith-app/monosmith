import { describe, expect, it } from 'vitest'
import { pickArguments, updateArguments } from './update-arguments.policy.ts'

describe('updateArguments', () => {
  it('names one package', () => {
    expect(updateArguments('eslint')).toEqual(['up', 'eslint'])
    expect(updateArguments('@types/node')).toEqual(['up', '@types/node'])
    expect(updateArguments('golang.org/x/text')).toEqual(['up', 'golang.org/x/text'])
  })

  it('refuses what could be an option or is not a name', () => {
    expect(updateArguments('--yes')).toBeUndefined()
    expect(updateArguments('-y')).toBeUndefined()
    expect(updateArguments('a b')).toBeUndefined()
    expect(updateArguments('')).toBeUndefined()
  })
})

describe('pickArguments', () => {
  it('lets the person pick in the terminal', () => {
    expect(pickArguments()).toEqual(['up'])
  })
})
