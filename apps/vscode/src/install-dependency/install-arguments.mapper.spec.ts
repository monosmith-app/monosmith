import { describe, expect, it } from 'vitest'
import { installArguments, parsePackages } from './install-arguments.mapper.ts'

describe('installArguments', () => {
  it('names a library with workspace: so smith does not ask which is meant', () => {
    expect(installArguments({ project: 'web', source: 'workspace', library: 'core', packages: [], dev: false })).toEqual(['install', '-w', 'web', 'workspace:core'])
  })

  it('passes the range the person chose', () => {
    expect(installArguments({ project: 'web', source: 'workspace', library: 'core', range: '*', packages: [], dev: false })).toEqual(['install', '-w', 'web', 'workspace:core', '--range', '*'])
  })

  it('sends registry packages with --registry and -D when asked', () => {
    expect(installArguments({ project: 'web', source: 'registry', packages: ['zod', 'core'], dev: true })).toEqual(['install', '-w', 'web', '-D', '--registry', 'zod', 'core'])
  })
})

describe('parsePackages', () => {
  it('splits on spaces and commas', () => {
    expect(parsePackages(' zod, react@19  @types/node ')).toEqual(['zod', 'react@19', '@types/node'])
  })
})
