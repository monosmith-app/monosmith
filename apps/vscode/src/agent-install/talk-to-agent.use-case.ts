/** What talking to the agent needs from the editor. */
export interface TalkDeps {
  /** The Claude Code command is available. */
  hasClaude:       () => Promise<boolean>
  /** The workspace has `.claude/agents/smith.md`. */
  hasAgentFile:    () => boolean
  /** Offers to install the agent; true when the files are there afterwards. */
  offerInstall:    () => Promise<boolean>
  /** Opens a terminal that runs `smith agent`. */
  openAgent:       () => void
  /** Says how to install Claude Code. */
  explainNoClaude: () => Promise<void>
}

/** How asking for the agent ended. */
export type TalkOutcome = 'started' | 'no-claude' | 'no-agent'

/**
 * Starts the agent in a terminal when Claude Code is there. Without Claude
 * Code it explains how to get it, and without the agent files it offers them first.
 */
export async function talkToAgent (deps: TalkDeps): Promise<TalkOutcome> {
  if (!await deps.hasClaude()) {
    await deps.explainNoClaude()

    return 'no-claude'
  }
  if (!deps.hasAgentFile() && !await deps.offerInstall()) return 'no-agent'
  deps.openAgent()

  return 'started'
}
