import { describe, expect, it } from 'vitest'
import { splitNoticeButtons } from './notice-buttons.mapper.ts'

const button = (label: string) => ({ label })

describe('splitNoticeButtons', () => {
  it('shows up to three buttons as they are', () => {
    expect(splitNoticeButtons([button('a'), button('b'), button('c')])).toEqual({ shown: [button('a'), button('b'), button('c')], more: [] })
  })

  it('keeps two buttons and puts the rest behind "More…" when there are four', () => {
    const split = splitNoticeButtons([button('Install tools'), button('Open Health'), button('Commit anyway'), button('Keep uncommitted')])
    expect(split.shown.map((item) => item.label)).toEqual(['Install tools', 'Open Health'])
    expect(split.more.map((item) => item.label)).toEqual(['Commit anyway', 'Keep uncommitted'])
  })
})
