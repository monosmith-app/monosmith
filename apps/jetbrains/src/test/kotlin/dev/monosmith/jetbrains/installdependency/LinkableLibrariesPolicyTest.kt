package dev.monosmith.jetbrains.installdependency

import dev.monosmith.jetbrains.workspace.SmithProjectContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class LinkableLibrariesPolicyTest {
    private fun node(
        name: String,
        kind: String,
        language: String = "node",
        deps: List<String> = emptyList(),
    ) = SmithProjectContract(name, name, kind, language, emptyList(), deps)

    private val nodes =
        listOf(
            node("web", "app", deps = listOf("ui")),
            node("ui", "lib", deps = listOf("core")),
            node("core", "lib"),
            node("pkg", "package"),
            node("tool", "lib", language = "go"),
        )

    @Test
    fun anAppCanUseALibraryOfTheSameLanguage() {
        assertTrue(LinkableLibrariesPolicy.mayLink(nodes, "web", "core"))
        assertTrue(LinkableLibrariesPolicy.mayLink(nodes, "web", "pkg"))
    }

    @Test
    fun neverItselfNeverAnAppNeverAnotherLanguage() {
        assertFalse(LinkableLibrariesPolicy.mayLink(nodes, "ui", "ui"))
        assertFalse(LinkableLibrariesPolicy.mayLink(nodes, "core", "web"))
        assertFalse(LinkableLibrariesPolicy.mayLink(nodes, "web", "tool"))
        assertFalse(LinkableLibrariesPolicy.mayLink(nodes, "web", "missing"))
    }

    @Test
    fun neverACycleDirectOrThroughAnotherLibrary() {
        assertFalse(LinkableLibrariesPolicy.mayLink(nodes, "core", "ui"))
        assertTrue(LinkableLibrariesPolicy.mayLink(nodes, "ui", "pkg"))
    }

    @Test
    fun listsTheLibrariesAProjectMayUseInTheOrderGiven() {
        assertEquals(listOf("ui", "core", "pkg"), LinkableLibrariesPolicy.linkable(nodes, "web").map { it.name })
        assertEquals(listOf("pkg"), LinkableLibrariesPolicy.linkable(nodes, "core").map { it.name })
    }
}
