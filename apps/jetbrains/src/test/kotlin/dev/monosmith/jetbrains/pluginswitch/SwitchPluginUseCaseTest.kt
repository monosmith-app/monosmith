package dev.monosmith.jetbrains.pluginswitch

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class SwitchPluginUseCaseTest {
    private val fake = FakePortsHarness()
    private val listed =
        """[{"name":"node","source":"built-in","roles":["language","toolchain"]},""" +
            """{"name":"go","source":"built-in","roles":["language"]},{"name":"deploy","source":"installed","roles":["deploy"]}]"""
    private val hides =
        "✖ switching off hides projects: node finds 2 projects (web, utils)\n" +
            "  → smith plugin disable node --force switches it off anyway\n"

    init {
        fake.answers["plugin list --json"] = fake.ok(listed)
    }

    @Test
    fun onlyTheBuiltInPluginsAreOfferedWithTheirRoles() {
        fake.ones.add("go")
        SwitchPluginUseCase.run(fake.ports, on = false)
        assertEquals(listOf("node", "go"), fake.rowsSeen.single().map { it.name })
        val rows = fake.rowsSeen.single()
        assertEquals("language, toolchain", rows.first().detail)
    }

    @Test
    fun turnsAPluginOnAfterTheCommandIsShown() {
        fake.ones.add("go")
        fake.plans.add(fake.yes())
        assertTrue(SwitchPluginUseCase.run(fake.ports, on = true))
        assertEquals("smith plugin enable go", fake.plansSeen.single())
        assertEquals(listOf("plugin", "enable", "go"), fake.calls.last())
        assertEquals("go is on. Nothing is committed.", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun turnsAPluginOffWhenNoProjectUsesIt() {
        fake.ones.add("go")
        fake.plans.add(fake.yes())
        assertTrue(SwitchPluginUseCase.run(fake.ports, on = false))
        assertEquals(listOf("plugin", "disable", "go"), fake.calls.last())
        assertEquals(1, fake.plansSeen.size)
    }

    @Test
    fun whenProjectsUseThePluginSmithsWarningIsShownAndOnlyASecondYesForcesIt() {
        fake.answers["plugin disable node"] = fake.fail(hides)
        fake.answers["plugin disable node --force"] = fake.ok("")
        fake.ones.add("node")
        fake.plans.addAll(listOf(fake.yes(), fake.yes()))
        assertTrue(SwitchPluginUseCase.run(fake.ports, on = false))
        assertTrue(fake.plansSeen.last().contains("node finds 2 projects (web, utils)"))
        assertEquals(listOf("plugin", "disable", "node", "--force"), fake.calls.last())
        assertEquals("node is off. Nothing is committed.", fake.infos.single())
    }

    @Test
    fun cancelAtTheSecondQuestionLeavesThePluginOn() {
        fake.answers["plugin disable node"] = fake.fail(hides)
        fake.ones.add("node")
        fake.plans.add(fake.yes())
        assertFalse(SwitchPluginUseCase.run(fake.ports, on = false))
        assertTrue(fake.calls.none { "--force" in it })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun anotherRefusalIsAWarningAndNeverForced() {
        fake.answers["plugin disable node"] = fake.fail("✖ node is not a built-in plugin\n")
        fake.ones.add("node")
        fake.plans.add(fake.yes())
        assertFalse(SwitchPluginUseCase.run(fake.ports, on = false))
        assertEquals("node was not turned off. node is not a built-in plugin", fake.warnings.single())
        assertEquals(1, fake.plansSeen.size)
    }

    @Test
    fun cancelAtTheCommandChangesNothing() {
        fake.ones.add("go")
        assertFalse(SwitchPluginUseCase.run(fake.ports, on = false))
        assertEquals(listOf(listOf("plugin", "list", "--json")), fake.calls)
    }

    @Test
    fun aListItCannotReadIsAWarningThatSaysToUpdateThePlugin() {
        fake.answers["plugin list --json"] = fake.ok("{}")
        assertFalse(SwitchPluginUseCase.run(fake.ports, on = true))
        assertTrue(fake.warnings.single().contains("Update the plugin"))
    }
}
