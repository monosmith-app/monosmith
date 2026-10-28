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

  it('is empty when no folder has a smith.toml and none looks adoptable', () => {
    expect(judgeViewState([])).toBe('empty')
    expect(judgeViewState([], 0)).toBe('empty')
  })

  it('is adoptable when no folder runs smith but one looks like a repository to bring in', () => {
    expect(judgeViewState([], 1)).toBe('adoptable')
  })

  it('shows the folders that run smith before any offer', () => {
    expect(judgeViewState(['ready'], 2)).toBe('ready')
    expect(judgeViewState(['missing'], 2)).toBe('missing')
  })
})
