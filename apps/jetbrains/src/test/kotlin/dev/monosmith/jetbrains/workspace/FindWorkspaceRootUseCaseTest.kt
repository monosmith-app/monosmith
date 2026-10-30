package dev.monosmith.jetbrains.workspace

import java.nio.file.Path
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class FindWorkspaceRootUseCaseTest {
    @Test
    fun findsTheNearestFolderUpWithSmithToml() {
        val root = Path.of("ws").toAbsolutePath()
        val found = FindWorkspaceRootUseCase.find(root.resolve("apps").resolve("api")) { it == root.resolve("smith.toml") }
        assertEquals(root, found)
    }

    @Test
    fun returnsNullOutsideAWorkspace() {
        assertNull(FindWorkspaceRootUseCase.find(Path.of("ws").toAbsolutePath()) { false })
    }
}
