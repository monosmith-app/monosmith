package dev.monosmith.jetbrains.convertfolder

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ConvertFolderTest {
    private val fake = FakePortsHarness()

    @Test
    fun theConservativeLevelNamesNoChoices() {
        assertEquals(
            listOf("convert", "--level", "conservative", "--dry-run"),
            ConvertArgumentsAlgorithm.arguments(ConvertChoicesContract(), apply = false),
        )
    }

    @Test
    fun theFullLevelNamesTheFourChoicesSoSmithAsksNothing() {
        assertEquals(
            listOf("convert", "--level", "full", "--lint", "smith", "--test", "jest", "--build", "keep", "--ci", "smith", "--yes"),
            ConvertArgumentsAlgorithm.arguments(ConvertChoicesContract("full", "smith", "jest", "keep", "smith"), apply = true),
        )
    }

    @Test
    fun theFirstAnswerOfEveryQuestionChangesTheLeastOrIsSmithsDefault() {
        assertEquals("conservative", ConvertArgumentsAlgorithm.levelRows.first().name)
        assertEquals("keep", ConvertArgumentsAlgorithm.lintRows.first().name)
        assertEquals("keep", ConvertArgumentsAlgorithm.testRows.first().name)
        assertEquals("smith", ConvertArgumentsAlgorithm.buildRows.first().name)
        assertEquals("smith", ConvertArgumentsAlgorithm.ciRows.first().name)
    }

    @Test
    fun theConservativeFlowShowsThePlanThenConvertsWithYes() {
        fake.answers["convert --level conservative --dry-run"] = fake.ok("would delete nothing\n")
        fake.ones.add("conservative")
        fake.plans.add(fake.yes())
        assertTrue(ConvertFolderUseCase.run(fake.ports) {})
        assertEquals("would delete nothing", fake.plansSeen.single())
        assertEquals(listOf("convert", "--level", "conservative", "--yes"), fake.calls.last())
    }

    @Test
    fun theFullFlowAsksTheFourChoicesAndShowsTheFilesItDeletes() {
        fake.answers["convert --level full --lint smith --test vitest --build smith --ci keep --dry-run"] =
            fake.ok("delete jest.config.js\n")
        fake.ones.addAll(listOf("full", "smith", "vitest", "smith", "keep"))
        fake.plans.add(fake.yes(commit = true))
        assertTrue(ConvertFolderUseCase.run(fake.ports) {})
        assertEquals("delete jest.config.js", fake.plansSeen.single())
        assertEquals(listOf("--ci", "keep", "--yes", "--commit"), fake.calls.last().takeLast(4))
    }

    @Test
    fun cancelLeavesTheWorkspaceAlone() {
        fake.ones.add("conservative")
        assertFalse(ConvertFolderUseCase.run(fake.ports) {})
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
