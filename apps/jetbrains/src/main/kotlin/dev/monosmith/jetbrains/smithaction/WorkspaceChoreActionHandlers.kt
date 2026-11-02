package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.agentfile.WriteAgentFileUseCase
import dev.monosmith.jetbrains.cachecontrol.CacheControlUseCase
import dev.monosmith.jetbrains.ciregenerate.RegenerateCiUseCase
import dev.monosmith.jetbrains.dockerhub.DockerDescriptionUseCase
import dev.monosmith.jetbrains.mcpinstall.McpInstallUseCase
import dev.monosmith.jetbrains.pipelinetemplate.AddPipelineUseCase
import dev.monosmith.jetbrains.pluginswitch.SwitchPluginUseCase
import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.settings.EditSettingsUseCase
import dev.monosmith.jetbrains.tauriapp.TauriAppUseCase

/** Turns a built-in plugin on; smith names nothing it hides. */
class PluginEnableActionHandler : SmithActionHandler({ ide, root -> SwitchPluginUseCase.start(ide, root, on = true) })

/** Turns a built-in plugin off; smith names the projects it hides, and a second yes forces it. */
class PluginDisableActionHandler : SmithActionHandler({ ide, root -> SwitchPluginUseCase.start(ide, root, on = false) })

/** Edits `dependencies.one_version` through the settings editor, with its reason and the plan. */
class OneVersionActionHandler : SmithActionHandler({ ide, root -> EditSettingsUseCase.startOne(ide, root, "dependencies.one_version") })

/** Shows the status of the local cache. */
class CacheStatusActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the cache status", listOf("cache", "status")) })

/** Checks the local cache for damage. */
class CacheVerifyActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the cache check", listOf("cache", "verify")) })

/** Removes what is broken in the local cache, after the consequence. */
class CacheRepairActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.REPAIR) })

/** Trims the local cache to its limit, after the plan. */
class CachePruneActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.PRUNE) })

/** Deletes the local cache, after the consequence. */
class CacheClearActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.CLEAR) })

/** Adds a ready-made pipeline, after the plan. */
class AddPipelineActionHandler : SmithActionHandler({ ide, root -> AddPipelineUseCase.start(ide, root) })

/** Regenerates the CI pipelines, after the diff of every file. */
class RegenerateCiActionHandler : SmithActionHandler({ ide, root -> RegenerateCiUseCase.start(ide, root) })

/** Makes the icons of a Tauri app from one PNG, after the plan. */
class TauriIconsActionHandler : SmithActionHandler({ ide, root -> TauriAppUseCase.startIcons(ide, root) })

/** Makes a Node app a Tauri app, after the plan; asks the bundle identifier. */
class TauriAddActionHandler : SmithActionHandler({ ide, root -> TauriAppUseCase.startAdd(ide, root) })

/** Sets the Docker Hub descriptions, after naming the repositories. */
class DockerDescriptionActionHandler : SmithActionHandler({ ide, root -> DockerDescriptionUseCase.start(ide, root) })

/** Says why a task would run again instead of coming from the cache. */
class WhyMissActionHandler : SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.startWhyMiss(ide, root) })

/** Lists the tools the workspace needs and the state of each. */
class ToolsListActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the tools", listOf("tools")) })

/** Shows the installed smith, the newest one and the settings of the workspace. */
class InfoActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the smith info", listOf("info")) })

/** Counts the files the projects keep as copies. */
class CopiesActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the copies between projects", listOf("copies")) })

/** Writes AGENTS.md from the rules of the workspace, after the plan. */
class AgentFileActionHandler : SmithActionHandler({ ide, root -> WriteAgentFileUseCase.start(ide, root) })

/** Connects smith to Claude Code or VS Code as an MCP server. */
class McpInstallActionHandler : SmithActionHandler({ ide, root -> McpInstallUseCase.start(ide, root) })
