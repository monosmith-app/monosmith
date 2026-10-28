import { describe, expect, it } from 'vitest'
import { talkToAgent, type TalkDeps } from './talk-to-agent.use-case.ts'

function fake (facts: { claude: boolean, file: boolean, installs?: boolean }) {
  const log: string[] = []
  const deps: TalkDeps = {
    hasClaude:    async () => facts.claude,
    hasAgentFile: () => facts.file,
    offerInstall: async () => {
      log.push('offer')

      return facts.installs ?? false
    },
    openAgent:       () => { log.push('open') },
    explainNoClaude: async () => { log.push('explain') },
  }

  return { log, deps }
}

describe('talkToAgent', () => {
  it('opens the agent when Claude Code and the agent file are there', async () => {
    const { log, deps } = fake({ claude: true, file: true })
    expect(await talkToAgent(deps)).toBe('started')
    expect(log).toEqual(['open'])
  })

  it('explains how to get Claude Code instead of opening a terminal that would fail', async () => {
    const { log, deps } = fake({ claude: false, file: true })
    expect(await talkToAgent(deps)).toBe('no-claude')
    expect(log).toEqual(['explain'])
  })

  it('offers the agent first when its file is missing, and opens it once installed', async () => {
    const { log, deps } = fake({ claude: true, file: false, installs: true })
    expect(await talkToAgent(deps)).toBe('started')
    expect(log).toEqual(['offer', 'open'])
  })

  it('stops when the person does not install the agent', async () => {
    const { log, deps } = fake({ claude: true, file: false, installs: false })
    expect(await talkToAgent(deps)).toBe('no-agent')
    expect(log).toEqual(['offer'])
  })
})
