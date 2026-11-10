package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.agentfile.WriteAgentFileUseCase

/** Writes AGENTS.md from the rules of the workspace, after the plan. */
class AgentFileActionHandler : SmithActionHandler({ ide, root -> WriteAgentFileUseCase.start(ide, root) })
