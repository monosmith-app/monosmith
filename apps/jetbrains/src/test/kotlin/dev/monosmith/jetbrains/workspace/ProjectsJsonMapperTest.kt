package dev.monosmith.jetbrains.workspace

import kotlin.test.Test
import kotlin.test.assertEquals

class ProjectsJsonMapperTest {
    @Test
    fun readsProjectsAndTheirTasks() {
        val json = """[{"name":"api","dir":"apps/api","kind":"app","language":"go","deps":[],"tasks":["build","test"]}]"""
        assertEquals(
            listOf(SmithProjectContract("api", "apps/api", "app", "go", listOf("build", "test"))),
            ProjectsJsonMapper.toProjects(json),
        )
    }

    @Test
    fun toleratesMissingOptionalFields() {
        assertEquals(
            listOf(SmithProjectContract("x", "x", "", "", emptyList())),
            ProjectsJsonMapper.toProjects("""[{"name":"x","dir":"x"}]"""),
        )
    }
}
