package dev.monosmith.jetbrains.smithcommand

import java.nio.file.Path
import kotlin.test.Test
import kotlin.test.assertEquals

class ResolveSmithExecutablePolicyTest {
    private val root = Path.of("ws")

    @Test
    fun prefersTheWorkspaceNodeInstall() {
        val found = ResolveSmithExecutablePolicy.resolve(root, windows = false) { true }
        assertEquals(
            root
                .resolve("node_modules")
                .resolve(".bin")
                .resolve("smith")
                .toString(),
            found,
        )
    }

    @Test
    fun usesTheSmithBinFolderNext() {
        val bin = root.resolve(".smith").resolve("bin").resolve("smith.exe")
        assertEquals(bin.toString(), ResolveSmithExecutablePolicy.resolve(root, windows = true) { it == bin })
    }

    @Test
    fun fallsBackToThePath() {
        assertEquals("smith", ResolveSmithExecutablePolicy.resolve(root, windows = false) { false })
    }
}
