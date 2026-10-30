package dev.monosmith.jetbrains.workspace

import java.nio.file.Path

/** Finds the MonoSmith workspace a folder belongs to: the nearest folder up with a smith.toml. */
object FindWorkspaceRootUseCase {
    fun find(
        start: Path,
        exists: (Path) -> Boolean,
    ): Path? = generateSequence(start) { it.parent }.firstOrNull { exists(it.resolve("smith.toml")) }
}
