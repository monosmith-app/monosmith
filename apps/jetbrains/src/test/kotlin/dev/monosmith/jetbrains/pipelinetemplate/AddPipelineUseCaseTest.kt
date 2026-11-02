package dev.monosmith.jetbrains.pipelinetemplate

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class AddPipelineUseCaseTest {
    private val fake = FakePortsHarness()
    private val listing =
        "e2e                                      github, azure  run every project's e2e task, one at a time\n" +
            "deploy-pages (--project <app>)           github         build an app and deploy its build output to GitHub Pages\n"

    init {
        fake.answers["pipeline"] = fake.ok(listing)
    }

    @Test
    fun readsTheNameTheProvidersAndWhetherAnAppIsNeeded() {
        val templates = PipelineTemplatesMapper.toTemplates(listing)
        assertEquals(listOf("e2e", "deploy-pages"), templates.map { it.name })
        assertEquals("github, azure", templates.first().providers)
        assertTrue(templates.last().needsApp)
        assertFalse(templates.first().needsApp)
        val rows = PipelineTemplatesMapper.rows(templates)
        assertTrue(rows.last().detail.endsWith("(asks for the app)"))
    }

    @Test
    fun aTemplateThatBuildsNoAppIsPlannedWithDryRunThenWritten() {
        fake.answers["pipeline e2e --dry-run"] = fake.ok("would write .github/workflows/e2e.yml\n")
        fake.ones.add("e2e")
        fake.plans.add(fake.yes())
        assertTrue(AddPipelineUseCase.run(fake.ports))
        assertEquals("would write .github/workflows/e2e.yml", fake.plansSeen.single())
        assertEquals(listOf("pipeline", "e2e", "--yes"), fake.calls.last())
        assertEquals("The e2e pipeline is written. Nothing is committed.", fake.infos.single())
    }

    @Test
    fun aDeployTemplateAsksForTheAppAndPassesItWithProject() {
        fake.ones.add("deploy-pages")
        fake.texts.add("web")
        fake.plans.add(fake.yes())
        assertTrue(AddPipelineUseCase.run(fake.ports))
        assertEquals(listOf("pipeline", "deploy-pages", "--project", "web", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelWritesNothing() {
        fake.ones.add("e2e")
        assertFalse(AddPipelineUseCase.run(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun anEmptyListIsAWarning() {
        fake.answers["pipeline"] = fake.ok("")
        assertFalse(AddPipelineUseCase.run(fake.ports))
        assertEquals("smith listed no pipeline. Update smith.", fake.warnings.single())
    }
}
