import { existsSync } from 'node:fs'
import { join } from 'node:path'
import type { CommandHost } from '../command-host/index.ts'
import type { FolderState } from '../folder-session/index.ts'
import { runDoctor } from '../health/index.ts'
import { openSmithTerminal } from '../smith-terminal/index.ts'
import { isWorkspaceTrusted } from '../workspace-trust/index.ts'
import { judgeAgentFiles, shouldOfferAgent, type AgentFilesState } from './agent-files.policy.ts'
import { confirmInstall, explainNoClaudeCode, offerAgent, tellAgentNeedsTrust, tellAgentOptedOut, tellAgentOptOutFailed, tellInstallOutcome } from './agent-prompts.client.ts'
import { hasClaudeCode, runSmith } from './command-line.client.ts'
import { installAgent } from './install-agent.use-case.ts'
import type { InstallScope } from './install-run.contract.ts'
import { optOutOfAgent } from './opt-out-agent.use-case.ts'
import { talkToAgent } from './talk-to-agent.use-case.ts'

const AGENT_ASKED_KEY = 'monosmith.agentOffered'
const AGENT_FILE = join('.claude', 'agents', 'smith.md')

/** What the agent commands need from the extension besides the host. */
export interface AgentFlowsDeps {
  /** What the doctor said about the agent files, by folder root; the flows write it. */
  agentFiles:     Map<string, AgentFilesState>
  /** Draws the views again. */
  sync:           () => void
  /** The editor's memory of what was already asked in this workspace. */
  workspaceState: { get: <T>(key: string) => T | undefined, update: (key: string, value: unknown) => PromiseLike<void> }
}

/** The commands of the agent menu. */
export interface AgentFlows {
  /** Asks the doctor whether the folder has the agent files, shows the line in the view, and offers them once. */
  checkAgentFiles: (state: FolderState, offer: boolean) => Promise<void>
  /** Installs the agent files or the MCP config in the folder of the command. */
  installFiles:    (scope: InstallScope, argument: unknown) => Promise<void>
  /** Starts a chat with the agent, installing it first when the person agrees. */
  talkToTheAgent:  (argument: unknown) => Promise<void>
}

/**
 * Builds the commands of the agent menu.
 * @param host - the folders and the Output channel of the extension
 * @param deps - the state the flows write and the views they draw
 * @returns the commands
 */
export function createAgentFlows (host: CommandHost, deps: AgentFlowsDeps): AgentFlows {
  const { agentFiles } = deps

  const declineAgent = async (state: FolderState): Promise<void> => {
    if (state.location === undefined) return
    const location = state.location
    const failure = await optOutOfAgent(async (args) => await runSmith(location.executable, args, state.root))
    if (failure !== undefined) {
      tellAgentOptOutFailed(failure)

      return
    }
    agentFiles.set(state.root, 'unknown')
    deps.sync()
    tellAgentOptedOut(state.name)
  }

  const installIn = async (state: FolderState, scope: InstallScope, confirmed: boolean): Promise<boolean> => {
    if (state.location === undefined) return false
    if (!isWorkspaceTrusted()) {
      tellAgentNeedsTrust()

      return false
    }
    const location = state.location
    const outcome = await installAgent({
      run:     async (args) => await runSmith(location.executable, args, state.root),
      confirm: async (preview) => await confirmInstall(preview, scope, state.name),
    }, scope, confirmed)
    host.note(`MonoSmith: install ${scope} in ${state.name}: ${outcome.kind}${outcome.kind === 'refused' ? ` (${outcome.message})` : ''}`)
    tellInstallOutcome(outcome, scope, state.root)
    if (outcome.kind === 'written' || outcome.kind === 'current') await checkAgentFiles(state, false)

    return outcome.kind === 'written' || outcome.kind === 'current'
  }

  const checkAgentFiles = async (state: FolderState, offer: boolean): Promise<void> => {
    const session = state.session
    if (session === undefined) return
    try {
      agentFiles.set(state.root, judgeAgentFiles(await runDoctor(session.connection)))
    } catch (error) {
      host.note(`MonoSmith: could not read the agent files check of ${state.name}: ${error instanceof Error ? error.message : String(error)}`)

      return
    }
    deps.sync()
    const asked = `${AGENT_ASKED_KEY}:${state.root}`
    if (!offer || !shouldOfferAgent({ state: agentFiles.get(state.root) ?? 'unknown', alreadyAsked: deps.workspaceState.get<boolean>(asked) === true })) return
    await deps.workspaceState.update(asked, true)
    const answer = await offerAgent(state.name)
    if (answer === 'install') await installIn(state, 'agent', false)
    else if (answer === 'never') await declineAgent(state)
  }

  return {
    checkAgentFiles,
    installFiles: async (scope, argument) => {
      const state = await host.folderFor(argument)
      if (state !== undefined) await installIn(state, scope, (argument as { confirmed?: unknown } | undefined)?.confirmed === true)
    },
    talkToTheAgent: async (argument) => {
      const state = await host.folderFor(argument)
      if (state?.location === undefined) return
      const location = state.location
      await talkToAgent({
        hasClaude:       hasClaudeCode,
        hasAgentFile:    () => existsSync(join(state.root, AGENT_FILE)),
        offerInstall:    async () => await installIn(state, 'agent', false) && existsSync(join(state.root, AGENT_FILE)),
        openAgent:       () => { openSmithTerminal(location.executable, ['agent'], state.root) },
        explainNoClaude: explainNoClaudeCode,
      })
    },
  }
}
