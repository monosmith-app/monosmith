package dev.monosmith.jetbrains.newworkspace

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NewWorkspaceUseCaseTest {
    private val fake = FakePortsHarness()
    private val questions =
        """
        [{"key":"ci","prompt":"CI?","default":"github","many":true,"options":[{"value":"github","label":"GitHub Actions"},{"value":"azure","label":"Azure Pipelines"}]},
         {"key":"language","prompt":"Language?","default":"later","options":[{"value":"later","label":"Decide later"},{"value":"go","label":"Go"}]},
         {"key":"architecture","prompt":"Layout?","default":"vertical-slices","options":[{"value":"vertical-slices","label":"vertical-slices"}]},
         {"key":"agent","prompt":"Agent?","default":"yes","options":[{"value":"yes","label":"Yes"},{"value":"no","label":"No"}]},
         {"key":"devcontainer","prompt":"Dev container?","default":"yes","options":[{"value":"yes","label":"Yes"},{"value":"no","label":"No"}]},
         {"key":"release-branch","prompt":"Branch?","default":"main","advanced":true}]
        """.trimIndent()

    init {
        fake.answers["new --questions --json"] = fake.ok(questions)
        fake.answers["new acme --here --yes"] = fake.ok("Would create smith.toml\n")
    }

    @Test
    fun theDefaultsCreateAfterTheFilesAreShown() {
        fake.ones.addAll(listOf("defaults"))
        fake.texts.add("acme")
        fake.plans.add(fake.yes())
        assertTrue(NewWorkspaceUseCase.run(fake.ports, "acme"))
        assertTrue(fake.plansSeen.single().contains("Would create smith.toml"))
        assertTrue(fake.calls.any { "--dry-run" in it })
        assertFalse("--dry-run" in fake.calls.last())
        assertEquals("--devcontainer", fake.calls.last().last())
    }

    @Test
    fun customiseAsksTheQuestionsOfSmithAndNotTheAdvancedOnes() {
        fake.ones.addAll(listOf("customise", "go", "vertical-slices", "no", "yes", "no"))
        fake.manies.add(listOf("github", "azure"))
        fake.texts.add("acme")
        fake.plans.add(fake.yes())
        assertTrue(NewWorkspaceUseCase.run(fake.ports, "acme"))
        val last = fake.calls.last()
        assertTrue("github,azure" in last && "go" in last && "--no-agent" in last && "--devcontainer" in last && "--no-commit" in last)
        assertFalse(fake.rowsSeen.any { rows -> rows.any { it.name == "main" } })
    }

    @Test
    fun anOldSmithThatCannotListItsQuestionsCreatesNothing() {
        fake.answers["new --questions --json"] = fake.fail("✖ unknown flag: --questions")
        assertFalse(NewWorkspaceUseCase.run(fake.ports, "acme"))
        assertTrue(fake.warnings.single().contains("Update smith"))
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun cancelWritesNothing() {
        fake.ones.addAll(listOf("defaults"))
        fake.texts.add("acme")
        assertFalse(NewWorkspaceUseCase.run(fake.ports, "acme"))
        assertTrue(fake.calls.none { it.first() == "new" && "--dry-run" !in it && "--questions" !in it })
    }
}
