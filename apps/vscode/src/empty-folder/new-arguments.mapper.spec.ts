import { describe, expect, it } from 'vitest'
import { toNewArguments } from './new-arguments.mapper.ts'

describe('toNewArguments', () => {
  it('runs smith new in the folder, interactively', () => {
    expect(toNewArguments()).toEqual(['new', '--here'])
  })
})
