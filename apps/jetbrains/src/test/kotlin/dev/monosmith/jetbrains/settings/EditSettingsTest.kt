package dev.monosmith.jetbrains.settings

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class EditSettingsTest {
    private val fake = FakePortsHarness()
    private val json =
        """
        [{"path":"workspace.confirm","type":"string","description":"Which plans ask.","reason":"every plan asks so you see what smith does","set":false,"enum":["always","risky"]},
         {"path":"workspace.exclude","type":"array","description":"Globs.","value":"[\"docs/**\",\"x/**\"]","set":true},
         {"path":"ci.devcontainer","type":"boolean","description":"A dev container.","value":"false","set":true},
         {"path":"workspace.module_prefix","type":"string","description":"Go prefix.","value":"\"github.com/acme\"","set":true},
         {"path":""}]
        """.trimIndent()
    private val settings = SettingsJsonMapper.toSettings(json)

    init {
        fake.answers["config list --json"] = fake.ok(json)
    }

    @Test
    fun readsTheValueAsAPersonTypesIt() {
        assertEquals("docs/**, x/**", SettingsJsonMapper.shown(settings[1]))
        assertEquals("github.com/acme", SettingsJsonMapper.shown(settings[3]))
        assertEquals("", SettingsJsonMapper.shown(settings[0]))
        assertEquals(4, settings.size)
    }

    @Test
    fun theListShowsTheValueOrDefault() {
        val rows = SettingsJsonMapper.rows(settings)
        assertEquals("(default)", rows.first().description)
        assertEquals("= docs/**, x/**", rows[1].description)
    }

    @Test
    fun anArrayIsOneArgumentPerItemAndAnythingElseIsOne() {
        assertEquals(
            listOf("config", "set", "workspace.exclude", "a/**", "b/**"),
            SettingsJsonMapper.setArguments(settings[1], "a/**, b/**"),
        )
        assertEquals(
            listOf("config", "set", "workspace.module_prefix", "github.com/x y"),
            SettingsJsonMapper.setArguments(settings[3], "github.com/x y"),
        )
        assertEquals(listOf("config", "unset", "workspace.confirm"), SettingsJsonMapper.unsetArguments(settings[0]))
    }

    @Test
    fun aBooleanAndAnEnumStartOnTheCurrentValue() {
        assertEquals(listOf("false", "true"), SettingsJsonMapper.booleanRows(settings[2]).map { it.name })
        assertEquals(listOf("always", "risky"), SettingsJsonMapper.enumRows(settings[0]).map { it.name })
        assertTrue(SettingsJsonMapper.words(settings[0]).contains("Why the default: every plan asks"))
    }

    @Test
    fun aSettingWithAFewValuesIsAListAndTheChangeIsShownFirst() {
        fake.ones.addAll(listOf("workspace.confirm", "risky"))
        fake.plans.add(fake.yes(commit = true))
        assertTrue(EditSettingsUseCase.run(fake.ports) {})
        assertEquals(listOf("config", "set", "workspace.confirm", "risky", "--commit"), fake.calls.last())
        assertEquals("smith config set workspace.confirm risky", fake.plansSeen.single())
    }

    @Test
    fun theCommitBoxIsUntickedAndAddsCommitOnlyWhenTicked() {
        fake.ones.addAll(listOf("workspace.confirm", "risky"))
        fake.plans.add(fake.yes())
        EditSettingsUseCase.run(fake.ports) {}
        assertFalse("--commit" in fake.calls.last())
    }

    @Test
    fun aSettingThatIsSetCanGoBackToItsDefault() {
        fake.ones.addAll(listOf("workspace.module_prefix", "reset"))
        fake.plans.add(fake.yes())
        assertTrue(EditSettingsUseCase.run(fake.ports) {})
        assertEquals(listOf("config", "unset", "workspace.module_prefix"), fake.calls.last())
    }

    @Test
    fun aTextSettingAsksForTheNewValueWithItsCurrentOne() {
        fake.ones.addAll(listOf("workspace.module_prefix", "change"))
        fake.texts.add("github.com/new")
        fake.plans.add(fake.yes())
        EditSettingsUseCase.run(fake.ports) {}
        assertEquals(listOf("config", "set", "workspace.module_prefix", "github.com/new"), fake.calls.last())
    }

    @Test
    fun cancelChangesNothing() {
        fake.ones.addAll(listOf("workspace.confirm", "risky"))
        assertFalse(EditSettingsUseCase.run(fake.ports) {})
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun theOneVersionSettingIsEditedDirectlyWithoutThePickFromTheList() {
        fake.answers["config list --json"] =
            fake.ok("""[{"path":"dependencies.one_version","type":"boolean","description":"One version.","value":"false","set":true}]""")
        fake.ones.addAll(listOf("change", "true"))
        fake.plans.add(fake.yes())
        assertTrue(EditSettingsUseCase.run(fake.ports, "dependencies.one_version") {})
        assertEquals(listOf("false", "true"), fake.rowsSeen.last().map { it.name })
        assertEquals("smith config set dependencies.one_version true", fake.plansSeen.single())
        assertEquals(listOf("config", "set", "dependencies.one_version", "true"), fake.calls.last())
        assertEquals(2, fake.rowsSeen.size)
    }

    @Test
    fun aSmithWithoutThatSettingIsAWarningAndChangesNothing() {
        assertFalse(EditSettingsUseCase.run(fake.ports, "dependencies.one_version") {})
        assertEquals("No setting was changed. This smith has no setting dependencies.one_version. Update smith.", fake.warnings.single())
        assertEquals(0, fake.changedCount)
    }
}
