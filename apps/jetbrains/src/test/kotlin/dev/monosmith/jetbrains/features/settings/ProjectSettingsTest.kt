package dev.monosmith.jetbrains.features.settings

import dev.monosmith.jetbrains.features.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ProjectSettingsTest {
    private val fake = FakePortsHarness()
    private val json =
        """
        {"name":"web","declared":false,"manifest":"apps/web/package.json",
         "facts":[{"path":"version","value":"1.2.0","description":"The version its manifest records.","source":"apps/web/package.json"}],
         "settings":[{"path":"projects.web.tag_prefix","type":"string","description":"Replaces the release tag prefix.","value":"\"web-v\"","set":true},
                     {"path":"projects.web.bare_tags","type":"boolean","description":"Tag releases with the bare version.","set":false}]}
        """.trimIndent()

    init {
        fake.answers["config list -p web --json"] = fake.ok(json)
    }

    @Test
    fun theReadOnlyValuesComeFirstWithTheirFileThenEachSettingByItsName() {
        val rows = ProjectSettingsUseCase.rows(SettingsJsonMapper.toProjectView(json))
        assertEquals(listOf("version", "tag_prefix", "bare_tags"), rows.map { it.label })
        assertTrue(rows.first().detail.contains("apps/web/package.json"))
        assertEquals("= web-v", rows[1].description)
    }

    @Test
    fun aReadOnlyValueSaysWhereToChangeItAndWritesNothing() {
        fake.ones.add("fact:version")
        assertFalse(ProjectSettingsUseCase.run(fake.ports, "web") {})
        assertTrue(fake.infos.single().contains("version comes from apps/web/package.json"))
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun aSettingIsWrittenAtItsProjectPath() {
        fake.ones.addAll(listOf("projects.web.bare_tags", "true"))
        fake.plans.add(fake.yes())
        assertTrue(ProjectSettingsUseCase.run(fake.ports, "web") {})
        assertEquals(listOf("config", "set", "projects.web.bare_tags", "true"), fake.calls.last())
    }

    @Test
    fun goingBackToTheDefaultRemovesTheLine() {
        fake.ones.addAll(listOf("projects.web.tag_prefix", "reset"))
        fake.plans.add(fake.yes())
        assertTrue(ProjectSettingsUseCase.run(fake.ports, "web") {})
        assertEquals(listOf("config", "unset", "projects.web.tag_prefix"), fake.calls.last())
    }
}
