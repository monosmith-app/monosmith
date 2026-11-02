package dev.monosmith.jetbrains.composestack

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ComposeStackAlgorithmTest {
    private fun item(
        name: String,
        kind: String = "choice",
        axis: String = "framework",
        layer: String = "framework",
        language: String = "node",
        many: Boolean = false,
        title: String = name,
        aliases: List<String> = emptyList(),
        isDefault: Boolean = false,
        suggested: Boolean = false,
    ) = ComposeItemContract(
        "$language/$name",
        kind,
        name,
        title,
        "the $name",
        aliases,
        layer,
        axis,
        many,
        language,
        suggested,
        isDefault,
        "",
        "",
    )

    private val items =
        listOf(
            item("node", kind = "language", axis = "", layer = "language", language = "node"),
            item("go", kind = "language", axis = "", layer = "language", language = "go"),
            item("react", title = "React"),
            item("vue", title = "Vue"),
            item("electron", title = "Electron", axis = "framework", layer = "shell", aliases = listOf("desktop")),
            item("tauri", title = "Tauri", axis = "shell", layer = "shell"),
            item("eslint", axis = "lint", layer = "lint", many = true),
            item("prettier", axis = "lint", layer = "lint", many = true),
        )

    private fun settle(
        previous: List<String>,
        chosen: List<String>,
    ) = ComposeStackAlgorithm.settleRivals(items, previous, chosen)

    @Test
    fun aNewFrameworkReplacesTheOneTickedBeforeLikeARadioButton() {
        val (ids, note) = settle(listOf("node/react"), listOf("node/react", "node/vue"))
        assertEquals(listOf("node/vue"), ids)
        assertEquals("Vue replaces React: a project has one framework.", note)
    }

    @Test
    fun aShellThatAnswersTheFrameworkQuestionReplacesAFramework() {
        val (ids, _) = settle(listOf("node/react"), listOf("node/react", "node/electron"))
        assertEquals(listOf("node/electron"), ids)
    }

    @Test
    fun aDifferentAxisIsKeptAndANewLanguageReplacesTheOldOne() {
        assertEquals(listOf("node/react", "node/tauri"), settle(listOf("node/react"), listOf("node/react", "node/tauri")).first)
        assertEquals(listOf("go/go"), settle(listOf("node/node"), listOf("node/node", "go/go")).first)
    }

    @Test
    fun twoChoicesOfAnAxisThatAllowsManyStayTogether() {
        val (ids, note) = settle(listOf("node/eslint"), listOf("node/eslint", "node/prettier"))
        assertEquals(listOf("node/eslint", "node/prettier"), ids)
        assertEquals("", note)
    }

    @Test
    fun noRivalMeansNoNote() {
        assertEquals(listOf("node/react") to "", settle(emptyList(), listOf("node/react")))
    }

    @Test
    fun buildsTheArgumentsOfSmithCompose() {
        assertEquals(listOf("compose", "--json"), ComposeStackAlgorithm.arguments(null, emptyList()))
        assertEquals(
            listOf("compose", "--json", "--lang", "node", "--with", "react,electron", "--kind", "app", "--name", "desk"),
            ComposeStackAlgorithm.arguments("node", listOf("react", "electron"), "app", "desk"),
        )
    }

    @Test
    fun searchFindsByAnAliasAndByAWordPrefix() {
        val electron = items.first { it.name == "electron" }
        assertTrue(ComposeStackAlgorithm.matches(electron, "desktop"))
        assertTrue(ComposeStackAlgorithm.matches(electron, "elec"))
        assertTrue(ComposeStackAlgorithm.matches(electron, "  "))
        assertFalse(ComposeStackAlgorithm.matches(electron, "svelte"))
    }

    @Test
    fun namesTheStackInWords() {
        assertEquals("React + Electron (node)", ComposeStackAlgorithm.stackSummary(items, "node", listOf("react", "electron")))
        assertEquals("React, Electron", ComposeStackAlgorithm.tickedTitles(items, listOf("node/electron", "node/react")))
    }

    @Test
    fun wordsThePlanWithTheFilesTheToolsAndTheNotes() {
        val plan =
            ComposePlanContract(
                command = "smith add app desk --lang node --with react,electron",
                argv = listOf("smith", "add"),
                files = listOf(ComposeFileContract("create", "apps/desk/package.json", "the manifest")),
                tools = listOf(ComposeToolContract("electron", "node", "Node.js", "missing", "smith tools install node")),
                notes = listOf("Run it in the folder."),
                recipeArgv = emptyList(),
            )
        val text = ComposeStackAlgorithm.describePlan(plan)
        assertTrue(text.startsWith("Run: smith add app desk"))
        assertTrue(text.contains("  create apps/desk/package.json (the manifest)"))
        assertTrue(text.contains("Node.js (missing, for electron): smith tools install node"))
        assertTrue(text.endsWith("Run it in the folder."))
        assertEquals(listOf("node"), ComposeStackAlgorithm.toolsToInstall(plan).map { it.name })
    }

    @Test
    fun markWhatSitsOnAnotherAxisAndWhatIsTheDefault() {
        val (label, where, _) = ComposeStackAlgorithm.labels(item("electron", title = "Electron", layer = "shell", isDefault = true))
        assertEquals("Electron (default)", label)
        assertEquals("node · shell (the framework choice)", where)
    }
}
