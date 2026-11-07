package dev.monosmith.jetbrains.runacross

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import dev.monosmith.jetbrains.workspace.AffectedJsonMapper
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class RunAcrossTest {
    private val fake = FakePortsHarness()
    private val projects =
        listOf(
            SmithProjectContract("web", "apps/web", "app", "node", listOf("build", "test", "start")),
            SmithProjectContract("api", "apps/api", "app", "go", listOf("build", "test")),
            SmithProjectContract("docs", "docs", "app", "node", listOf("build")),
        )

    @Test
    fun aTaskIsOneWordAndNeverAFlag() {
        assertEquals(listOf("run", "build"), RunAcrossAlgorithm.arguments("build", affected = false))
        assertEquals(listOf("run", "build", "--affected"), RunAcrossAlgorithm.arguments("build", affected = true))
        assertNull(RunAcrossAlgorithm.arguments("--all", affected = false))
        assertNull(RunAcrossAlgorithm.arguments("two words", affected = false))
        assertNull(RunAcrossAlgorithm.arguments("", affected = false))
    }

    @Test
    fun theTasksAreCountedAcrossTheProjectsOrOnlyTheAffectedOnes() {
        val all = RunAcrossAlgorithm.taskRows(projects, emptySet(), affectedOnly = false)
        assertEquals(listOf("build", "test", "start"), all.map { it.name })
        assertEquals("3 projects", all.first().description)
        val some = RunAcrossAlgorithm.taskRows(projects, setOf("api"), affectedOnly = true)
        assertEquals(listOf("build", "test"), some.map { it.name })
        assertEquals("1 project", some.first().description)
    }

    @Test
    fun readsTheAffectedProjects() {
        val affected = AffectedJsonMapper.toAffected("""{"projects":["web","api"],"all":false,"reason":"files changed"}""")
        assertEquals(setOf("web", "api"), affected.projects)
        assertFalse(affected.all)
    }

    @Test
    fun runAffectedAsksOnlyForTheTasksOfTheAffectedProjectsAndRunsWithTheFlag() {
        fake.answers["affected --json"] = fake.ok("""{"projects":["api"],"all":false}""")
        fake.ones.add("test")
        assertTrue(RunAcrossUseCase.run(fake.ports, projects, affected = true))
        assertEquals(listOf("run", "test", "--affected"), fake.calls.last())
        assertTrue(fake.shown.last())
        assertEquals(listOf("build", "test"), fake.rowsSeen.single().map { it.name })
    }

    @Test
    fun runInAllProjectsNeverAsksWhichAreAffected() {
        fake.ones.add("build")
        assertTrue(RunAcrossUseCase.run(fake.ports, projects, affected = false))
        assertEquals(listOf(listOf("run", "build")), fake.calls)
    }

    @Test
    fun noAffectedProjectMeansNoQuestionAndNoRun() {
        fake.answers["affected --json"] = fake.ok("""{"projects":[],"all":false}""")
        assertFalse(RunAcrossUseCase.run(fake.ports, projects, affected = true))
        assertEquals("No project is affected by the changes.", fake.infos.single())
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun aFailedTaskIsAWarningWithSmithsWords() {
        fake.answers["run build"] = fake.fail("× build failed in web")
        fake.ones.add("build")
        RunAcrossUseCase.run(fake.ports, projects, affected = false)
        assertEquals("build did not finish. build failed in web", fake.warnings.single())
    }
}
