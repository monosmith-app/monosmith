package dev.monosmith.jetbrains.dependencies

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class UpdatesTest {
    private val fake = FakePortsHarness()
    private val json =
        """
        {"updates":[
          {"plugin":"node","name":"zod","current":"3.22.0","latest":"3.23.1","level":"minor","sites":[{"project":"web"},{"project":"api"},{"project":"web"}]},
          {"plugin":"node","name":"react","current":"18.2.0","latest":"19.0.0","level":"major","sites":[{"project":"web"}]},
          {"plugin":"go","name":"yaml","current":"1.0.0","latest":"1.0.1","level":"patch","sites":[{"project":"api"}]}],
         "tools":[{"name":"Go","pinned":"1.26","latest":"1.27.3","newer":true},{"name":"Node","pinned":"22","latest":"22","newer":false}]}
        """.trimIndent()

    @Test
    fun readsTheUpdatesWithTheirProjectsAndOnlyTheNewerTools() {
        val report = UpdatesJsonMapper.toReport(json)
        assertEquals(3, report.updates.size)
        assertEquals(listOf("web", "api"), report.updates.first().projects)
        assertEquals(listOf("Go"), report.tools.map { it.name })
    }

    @Test
    fun theRegistryPageOfAnUpdateIsReadAndNamedInThePlanWhenKnown() {
        val withPage =
            UpdatesJsonMapper.toReport(
                """
                {"updates":[{"plugin":"node","name":"zod","current":"3.22.0","latest":"3.23.1","level":"minor","url":"https://www.npmjs.com/package/zod","sites":[{"project":"web"}]}],"tools":[]}
                """.trimIndent(),
            )
        assertEquals("https://www.npmjs.com/package/zod", withPage.updates.single().url)
        assertTrue(UpdatesJsonMapper.describe(withPage.updates).contains("registry: https://www.npmjs.com/package/zod"))
        val without = UpdatesJsonMapper.toReport(json)
        assertEquals("", without.updates.first().url)
        assertFalse(UpdatesJsonMapper.describe(without.updates).contains("registry:"))
    }

    @Test
    fun groupsByRiskWithTheRiskiestFirst() {
        val groups = UpdatesJsonMapper.byRisk(UpdatesJsonMapper.toReport(json).updates)
        assertEquals(listOf("major", "minor", "patch"), groups.map { it.first })
        assertEquals("zod 3.22.0 → 3.23.1 (2 projects)", UpdatesJsonMapper.line(groups[1].second.single()))
    }

    @Test
    fun aMajorUpdateIsNeverTickedAtFirst() {
        val rows = UpdatesJsonMapper.rows(UpdatesJsonMapper.toReport(json).updates)
        assertEquals(listOf("yaml", "zod", "react"), rows.map { it.name })
        assertEquals(listOf(true, true, false), rows.map { it.picked })
    }

    @Test
    fun theCommandLinesTakeMajorOnlyWhenAMajorIsPicked() {
        assertEquals(listOf("up", "--yes"), UpdatesJsonMapper.arguments(emptyList(), false))
        assertEquals(listOf("up", "zod", "react", "--major", "--yes"), UpdatesJsonMapper.arguments(listOf("zod", "react"), true))
        assertEquals(listOf("up", "--align", "--prefer", "highest", "--dry-run"), UpdatesJsonMapper.alignArguments("highest", false))
        assertEquals(listOf("up", "--align", "--prefer", "highest", "--yes"), UpdatesJsonMapper.alignArguments("highest", true))
    }

    @Test
    fun theSafeUpdatesAreTheDefaultAndNeverCarryMajorOrACommit() {
        fake.answers["up --json"] = fake.ok(json)
        fake.ones.add("safe")
        fake.plans.add(fake.yes())
        assertTrue(UpdateDependenciesUseCase.update(fake.ports) {})
        assertEquals(listOf("up", "--yes"), fake.calls.last())
        assertTrue(fake.plansSeen.single().contains("zod 3.22.0 → 3.23.1 (2 projects) · minor"))
        assertFalse(fake.plansSeen.single().contains("react"))
    }

    @Test
    fun aPickedMajorNeedsTheMajorFlagAndTheCommitBoxStartsUnticked() {
        fake.answers["up --json"] = fake.ok(json)
        fake.ones.add("choose")
        fake.manies.add(listOf("react"))
        fake.plans.add(fake.yes(commit = true))
        UpdateDependenciesUseCase.update(fake.ports) {}
        assertEquals(listOf("up", "react", "--major", "--yes", "--commit"), fake.calls.last())
    }

    @Test
    fun nothingToUpdateSaysSoAndAsksNothing() {
        fake.answers["up --json"] = fake.ok("""{"updates":[],"tools":[]}""")
        assertFalse(UpdateDependenciesUseCase.update(fake.ports) {})
        assertTrue(fake.infos.single().startsWith("Every dependency is up to date."))
    }

    @Test
    fun alignAsksWhichVersionWinsThenShowsThePlan() {
        fake.answers["up --align --prefer installed --dry-run"] = fake.ok("axios 1.6 -> 1.7 in api\n")
        fake.ones.add("installed")
        fake.plans.add(fake.yes())
        assertTrue(UpdateDependenciesUseCase.align(fake.ports) {})
        assertEquals("axios 1.6 -> 1.7 in api", fake.plansSeen.single())
        assertEquals(listOf("up", "--align", "--prefer", "installed", "--yes"), fake.calls.last())
        assertEquals(
            "installed",
            fake.rowsSeen
                .single()
                .first()
                .name,
        )
    }
}
