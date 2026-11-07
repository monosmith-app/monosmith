package dev.monosmith.jetbrains.pluginmarket

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class InstallPluginTest {
    private val fake = FakePortsHarness()
    private val found =
        """[{"name":"widget","newest":"1.2.0","description":"a widget deployer","roles":["deploy"],"languages":["go"],"license":"MIT"},{"description":"no name"}]"""

    @Test
    fun readsThePluginsThatHaveAName() {
        val plugins = PluginSearchMapper.toPlugins(found)
        assertEquals(listOf("widget"), plugins.map { it.name })
        assertEquals("widget 1.2.0", PluginSearchMapper.rows(plugins).single().label)
        assertEquals("go · deploy", PluginSearchMapper.rows(plugins).single().description)
    }

    @Test
    fun theWordsOfTheTermAreSeparateArguments() {
        assertEquals(listOf("plugin", "search", "deploy", "go", "--json"), PluginSearchMapper.searchArguments(" deploy  go "))
        assertEquals(listOf("plugin", "search", "--json"), PluginSearchMapper.searchArguments(""))
    }

    @Test
    fun trustAndPinAreOffUnlessAsked() {
        val plugin = PluginSearchMapper.toPlugins(found).single()
        assertEquals(listOf("plugin", "install", "widget@1.2.0"), PluginSearchMapper.installArguments(plugin, trust = false, pin = false))
        assertEquals(
            listOf("plugin", "install", "widget@1.2.0", "--trust", "--pin"),
            PluginSearchMapper.installArguments(plugin, trust = true, pin = true),
        )
        assertTrue(PluginSearchMapper.trustWords(plugin).contains("Installing does not trust it"))
    }

    @Test
    fun installsTheNewestVersionWithoutTrustWhenTheBoxesStayUnticked() {
        fake.answers["plugin search deploy --json"] = fake.ok(found)
        fake.texts.add("deploy")
        fake.ones.add("widget")
        fake.plans.add(PlanAnswerContract(false, listOf(false, false), -1))
        assertTrue(InstallPluginUseCase.run(fake.ports))
        assertEquals(listOf("plugin", "install", "widget@1.2.0"), fake.calls.last())
        assertTrue(fake.plansSeen.single().contains("smith plugin install widget@1.2.0"))
    }

    @Test
    fun tickedBoxesBecomeTheFlags() {
        fake.answers["plugin search --json"] = fake.ok(found)
        fake.texts.add("")
        fake.ones.add("widget")
        fake.plans.add(PlanAnswerContract(false, listOf(true, true), -1))
        InstallPluginUseCase.run(fake.ports)
        assertEquals(listOf("plugin", "install", "widget@1.2.0", "--trust", "--pin"), fake.calls.last())
    }

    @Test
    fun noMatchInstallsNothing() {
        fake.answers["plugin search"] = fake.ok("[]")
        fake.texts.add("zzz")
        assertFalse(InstallPluginUseCase.run(fake.ports))
        assertEquals("No plugin matches \"zzz\".", fake.infos.single())
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun aRefusedInstallIsAWarningWithSmithsWords() {
        fake.answers["plugin search"] = fake.ok(found)
        fake.answers["plugin install"] = fake.fail("× the archive does not match its SHA-256")
        fake.texts.add("")
        fake.ones.add("widget")
        fake.plans.add(PlanAnswerContract(false, listOf(false, false), -1))
        assertFalse(InstallPluginUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("widget was not installed. the archive does not match"))
    }
}
