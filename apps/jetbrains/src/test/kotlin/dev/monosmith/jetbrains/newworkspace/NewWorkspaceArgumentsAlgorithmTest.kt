package dev.monosmith.jetbrains.newworkspace

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NewWorkspaceArgumentsAlgorithmTest {
    private val questions =
        NewQuestionsMapper.toQuestions(
            """
            [{"key":"ci","prompt":"CI?","default":"github","flag":"ci","many":true,"options":[{"value":"github","label":"GitHub Actions","hint":"default"},{"value":"gitlab","label":"GitLab CI"}]},
             {"key":"language","prompt":"Language?","default":"later","flag":"language","options":[{"value":"later","label":"Decide later"},{"value":"node","label":"TypeScript / JavaScript"}]},
             {"key":"node-stack","prompt":"Node?","default":"node-jest","options":[{"value":"node-jest","label":"Jest"}],"when":{"language":"node"}},
             {"key":"architecture","prompt":"Layout?","default":"hexagonal","options":[{"value":"hexagonal","label":"hexagonal"}]},
             {"key":"agent","prompt":"Agent?","default":"no","options":[{"value":"no","label":"No"},{"value":"yes","label":"Yes"}]}]
            """.trimIndent(),
        )

    @Test
    fun theDefaultsAreWhatSmithSays() {
        val defaults = NewWorkspaceArgumentsAlgorithm.defaults("acme", questions)
        assertEquals(listOf("github"), defaults.ci)
        assertEquals("later", defaults.language)
        assertEquals("hexagonal", defaults.architecture)
        assertFalse(defaults.agent)
        assertTrue(defaults.devcontainer)
    }

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
    fun describesTheChoicesWithTheLabelsOfSmithsQuestions() {
        val lines = NewWorkspaceArgumentsAlgorithm.describe(NewWorkspaceChoicesContract("acme", language = "node"), "acme", questions)
        assertEquals("Name: acme (the name of this folder)", lines.first())
        assertTrue(lines.contains("CI: GitHub Actions"))
        assertTrue(lines.contains("Main language: TypeScript / JavaScript"))
        assertTrue(lines.contains("Tests and lint for Node projects: Jest"))
        assertTrue(lines.last().contains("yes, smith makes it"))
        assertFalse(
            NewWorkspaceArgumentsAlgorithm.describe(NewWorkspaceChoicesContract("acme"), "acme", questions).any {
                it.startsWith("Tests and lint")
            },
        )
    }

    @Test
    fun theYesNoQuestionStartsOnItsDefault() {
        assertEquals(listOf("yes", "no"), NewWorkspaceArgumentsAlgorithm.yesNoRows("a", "b", true).map { it.name })
        assertEquals(listOf("no", "yes"), NewWorkspaceArgumentsAlgorithm.yesNoRows("a", "b", false).map { it.name })
    }
}
