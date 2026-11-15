package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.mcpinstall.McpInstallUseCase

/** Connects smith to Claude Code or VS Code as an MCP server. */
class McpInstallActionHandler : SmithActionHandler({ ide, root -> McpInstallUseCase.start(ide, root) })
