import { describe, expect, it, vi } from 'vitest'
import { offerTabCompletion, type TabCompletionAnswer, type TabCompletionOfferDeps } from './offer-tab-completion.use-case.ts'
import type { TabCompletionMemory } from './tab-completion-offer.policy.ts'

const WANTED = { code: 0, stdout: '{"shell":"zsh","wanted":true,"answer":""}', stderr: '' }

function editor (answer: TabCompletionAnswer | undefined, overrides: Partial<TabCompletionOfferDeps> = {}): { deps: TabCompletionOfferDeps, kept: TabCompletionMemory[], install: ReturnType<typeof vi.fn> } {
  const kept: TabCompletionMemory[] = []
  const install = vi.fn()
  let asked = false
  const deps: TabCompletionOfferDeps = {
    recall:           () => undefined,
    remember:         async (value) => { kept.push(value) },
    askedThisSession: () => asked,
    markAsked:        () => { asked = true },
    status:           async () => WANTED,
    ask:              async () => answer,
    install,
    ...overrides,
  }

  return { deps, kept, install }
}

describe('offerTabCompletion', () => {
  it('"Set up" opens the terminal on the install command and is remembered; the editor writes nothing', async () => {
    const { deps, kept, install } = editor('setup')
    expect(await offerTabCompletion(deps)).toBe(true)
    expect(install).toHaveBeenCalledOnce()
    expect(kept).toEqual(['done'])
  })

  it('"Don\'t ask again" is remembered and installs nothing', async () => {
    const { deps, kept, install } = editor('never')
    await offerTabCompletion(deps)
    expect(kept).toEqual(['never'])
    expect(install).not.toHaveBeenCalled()
  })

  it('"Not now" and a closed message remember nothing but are not asked again in the session', async () => {
    for (const answer of ['later', undefined] as const) {
      const { deps, kept, install } = editor(answer)
      await offerTabCompletion(deps)
      expect(kept).toEqual([])
      expect(install).not.toHaveBeenCalled()
      expect(await offerTabCompletion(deps)).toBe(false)
    }
  })

  it('does not ask when smith says it is not wanted, fails, or is too old to answer', async () => {
    for (const stdout of ['{"shell":"zsh","wanted":false,"answer":"installed"}', 'unknown command']) {
      const ask = vi.fn()
      expect(await offerTabCompletion(editor('setup', { status: async () => ({ code: 0, stdout, stderr: '' }), ask }).deps)).toBe(false)
      expect(ask).not.toHaveBeenCalled()
    }
    const ask = vi.fn()
    expect(await offerTabCompletion(editor('setup', { status: async () => ({ code: 1, stdout: '', stderr: 'x' }), ask }).deps)).toBe(false)
    expect(ask).not.toHaveBeenCalled()
  })

  it('does not even ask smith after an answer that is remembered', async () => {
    const status = vi.fn()
    expect(await offerTabCompletion(editor('setup', { recall: () => 'never', status }).deps)).toBe(false)
    expect(status).not.toHaveBeenCalled()
  })
})
