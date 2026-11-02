package dev.monosmith.jetbrains.newworkspace

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NewWorkspaceArgumentsAlgorithmTest {
    @Test
    fun theDefaultsGiveOneFlagForEveryQuestionSoSmithAsksNothingMore() {
        assertEquals(
            listOf(
                "new",
                "acme",
                "--here",
                "--yes",
                "--ci",
                "github",
                "--language",
                "later",
                "--architecture",
                "vertical-slices",
                "--agent",
                "--devcontainer",
            ),
            NewWorkspaceArgumentsAlgorithm.arguments(NewWorkspaceChoicesContract("acme"), preview = false),
        )
    }

    @Test
    fun aNodeWorkspaceNamesItsStackAndAPreviewIsADryRun() {
        val choices =
            NewWorkspaceChoicesContract(
                "acme",
                ci = listOf("github", "gitlab"),
                language = "node",
                nodeStack = "node-vitest",
                agent = false,
                devcontainer = false,
                commit = false,
            )
        val args = NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = true)
        assertTrue("github,gitlab" in args && "--default-stack" in args && "node=node-vitest" in args)
        assertTrue("--no-agent" in args && "--no-devcontainer" in args && "--no-commit" in args)
        assertEquals("--dry-run", args.last())
        assertFalse("--agent" in args)
    }

    @Test
    fun describesTheChoicesAndSaysTheNameIsTheFoldersName() {
        val lines = NewWorkspaceArgumentsAlgorithm.describe(NewWorkspaceChoicesContract("acme", language = "node"), "acme")
        assertEquals("Name: acme (the name of this folder)", lines.first())
        assertTrue(lines.contains("Main language: TypeScript / JavaScript"))
        assertTrue(lines.contains("Tests and lint for Node projects: Jest"))
        assertTrue(lines.last().contains("yes, smith makes it"))
        assertFalse(
            NewWorkspaceArgumentsAlgorithm.describe(NewWorkspaceChoicesContract("acme"), "acme").any { it.startsWith("Tests and lint") },
        )
    }

    @Test
    fun theYesNoQuestionStartsOnItsDefault() {
        assertEquals(listOf("yes", "no"), NewWorkspaceArgumentsAlgorithm.yesNoRows("a", "b", true).map { it.name })
        assertEquals(listOf("no", "yes"), NewWorkspaceArgumentsAlgorithm.yesNoRows("a", "b", false).map { it.name })
    }

    @Test
    fun theFirstAnswerOfEveryListIsTheDefault() {
        listOf(
            NewWorkspaceArgumentsAlgorithm.languageRows,
            NewWorkspaceArgumentsAlgorithm.nodeStackRows,
            NewWorkspaceArgumentsAlgorithm.architectureRows,
        ).forEach {
            assertEquals("(default)", it.first().description)
        }
        assertEquals(listOf(true, false, false, false, false), NewWorkspaceArgumentsAlgorithm.ciRows.map { it.picked })
    }
}
