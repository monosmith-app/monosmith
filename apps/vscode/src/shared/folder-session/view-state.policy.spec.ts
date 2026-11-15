import { describe, expect, it } from 'vitest'
import { judgeViewState } from './view-state.policy.ts'

describe('judgeViewState', () => {
  it('is ready when any folder is ready', () => {
    expect(judgeViewState(['missing', 'ready', 'failed'])).toBe('ready')
  })

  it('is starting while a folder starts and none is ready', () => {
    expect(judgeViewState(['failed', 'starting'])).toBe('starting')
  })

  it('names trust before a missing smith, and that before a failure', () => {
    expect(judgeViewState(['failed', 'missing', 'untrusted'])).toBe('untrusted')
    expect(judgeViewState(['failed', 'missing'])).toBe('missing')
    expect(judgeViewState(['failed'])).toBe('failed')
  })

  it('is nofolder when nothing is open', () => {
    expect(judgeViewState([])).toBe('nofolder')
    expect(judgeViewState([], { adoptable: 0, empty: 0, unrecognised: 0 })).toBe('nofolder')
  })

  it('is adoptable when no folder runs smith but one looks like a repository to bring in', () => {
    expect(judgeViewState([], { adoptable: 1 })).toBe('adoptable')
  })

  it('is empty for an empty folder and unrecognised for one with files and no known tool', () => {
    expect(judgeViewState([], { empty: 1 })).toBe('empty')
    expect(judgeViewState([], { unrecognised: 1 })).toBe('unrecognised')
  })

  it('prefers a conversion to an empty folder, and that to a folder with unknown files', () => {
    expect(judgeViewState([], { adoptable: 1, empty: 1, unrecognised: 1 })).toBe('adoptable')
    expect(judgeViewState([], { empty: 1, unrecognised: 1 })).toBe('empty')
  })

  it('shows the folders that run smith before any offer', () => {
    expect(judgeViewState(['ready'], { adoptable: 2, empty: 1 })).toBe('ready')
    expect(judgeViewState(['missing'], { adoptable: 2, empty: 1 })).toBe('missing')
  })
})
