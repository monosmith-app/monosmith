package dev.monosmith.jetbrains.smithcommand

import java.nio.file.Path

/**
 * Which smith the plugin runs: the workspace's own install, then the one on the PATH, so the IDE,
 * the terminal and CI run the same version (docs/architecture/decisions/0025).
 */
object ResolveSmithExecutablePolicy {
    fun resolve(
        workspaceRoot: Path,
        windows: Boolean,
        exists: (Path) -> Boolean,
    ): String {
        val node = workspaceRoot.resolve("node_modules").resolve(".bin").resolve(if (windows) "smith.cmd" else "smith")
        val installed = workspaceRoot.resolve(".smith").resolve("bin").resolve(if (windows) "smith.exe" else "smith")
        return listOf(node, installed).firstOrNull(exists)?.toString() ?: "smith"
    }
}
