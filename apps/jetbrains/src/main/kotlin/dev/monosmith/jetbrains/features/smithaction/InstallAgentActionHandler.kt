package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** Installs the MCP server and the agent files for Claude Code, after the diffs. */
class InstallAgentActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.installAgent(ide, root) })
