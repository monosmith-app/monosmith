package dev.monosmith.jetbrains.convertfolder

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ConvertFolderTest {
    private val fake = FakePortsHarness()
    private val questions =
        """
        [{"key":"level","prompt":"How far?","default":"conservative","flag":"level","options":[{"value":"conservative","label":"conservative"},{"value":"full","label":"full"}]},
         {"key":"lint","prompt":"Lint with:","default":"keep","flag":"lint","when":{"level":"full"},"options":[{"value":"keep","label":"keep"},{"value":"smith","label":"smith"}]},
         {"key":"test","prompt":"Run tests with:","default":"keep","flag":"test","when":{"level":"full"},"options":[{"value":"keep","label":"keep"},{"value":"jest","label":"jest"},{"value":"vitest","label":"vitest"}]},
         {"key":"build","prompt":"Build with:","default":"smith","flag":"build","when":{"level":"full"},"options":[{"value":"smith","label":"smith"},{"value":"keep","label":"keep"}]},
         {"key":"ci","prompt":"CI pipeline:","default":"smith","flag":"ci","when":{"level":"full"},"options":[{"value":"smith","label":"smith"},{"value":"keep","label":"keep"}]}]
        """.trimIndent()

    init {
        fake.answers["convert --questions --json"] = fake.ok(questions)
    }

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

    @Test
    fun theConservativeLevelAsksOnlyTheLevelBecauseTheChoicesApplyToTheFullLevelOnly() {
        fake.ones.add("conservative")
        fake.plans.add(null)
        ConvertFolderUseCase.run(fake.ports) {}
        assertEquals(1, fake.rowsSeen.size)
        assertEquals(
            "conservative",
            fake.rowsSeen
                .single()
                .first()
                .name,
        )
    }

    @Test
    fun theFullLevelShowsEachQuestionWithSmithsDefaultFirst() {
        fake.ones.addAll(listOf("full", "keep", "keep", "smith", "smith"))
        fake.plans.add(null)
        ConvertFolderUseCase.run(fake.ports) {}
        assertEquals(listOf("conservative", "keep", "keep", "smith", "smith"), fake.rowsSeen.map { it.first().name })
    }

    @Test
    fun anOldSmithThatCannotListTheQuestionsConvertsNothing() {
        fake.answers["convert --questions --json"] = fake.fail("✖ unknown flag: --questions")
        assertFalse(ConvertFolderUseCase.run(fake.ports) {})
        assertTrue(fake.warnings.single().contains("Update smith"))
        assertEquals(1, fake.calls.size)
    }
}
