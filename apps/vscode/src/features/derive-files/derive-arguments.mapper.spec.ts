import { describe, expect, it } from 'vitest'
import { applyActionFor, deriveArguments } from './derive-arguments.mapper.ts'

describe('deriveArguments', () => {
  it('checks without writing', () => {
    expect(deriveArguments('check')).toEqual(['derive', '--check'])
  })

  it('applies everything with a plain derive', () => {
    expect(deriveArguments('apply')).toEqual(['derive', '--yes'])
  })

  it('explains a file', () => {
    expect(deriveArguments('explain', 'a/tsconfig.json')).toEqual(['derive', '--explain=a/tsconfig.json'])
  })

  it('replaces, adopts and ejects one file', () => {
    expect(deriveArguments('replace', 'a.json')).toEqual(['derive', '--replace=a.json', '--yes'])
    expect(deriveArguments('adopt', 'a.json')).toEqual(['derive', '--adopt=a.json', '--yes'])
    expect(deriveArguments('eject', 'a.json')).toEqual(['derive', '--eject=a.json', '--yes'])
  })

  it('never lets a file that starts with a dash become an option', () => {
    expect(deriveArguments('eject', '--yes')).toEqual(['derive', '--eject=--yes', '--yes'])
  })
})

describe('applyActionFor', () => {
  it('overwrites a file smith holds back, and writes the others', () => {
    expect(['drift', 'diverged', 'broken'].map((status) => applyActionFor(status as 'drift'))).toEqual(['replace', 'replace', 'replace'])
    expect(['stale', 'missing', 'unrecorded'].map((status) => applyActionFor(status as 'stale'))).toEqual(['apply', 'apply', 'apply'])
  })
})
