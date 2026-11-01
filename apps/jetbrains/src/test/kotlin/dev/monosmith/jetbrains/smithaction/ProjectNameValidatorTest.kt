package dev.monosmith.jetbrains.smithaction

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class ProjectNameValidatorTest {
    @Test
    fun acceptsKebabCase() {
        assertNull(ProjectNameValidator.problem("billing-api", emptyList()))
    }

    @Test
    fun refusesWhatSmithRefuses() {
        listOf("my app", "apps/web", "Web", "", "-web").forEach {
            assertNotNull(ProjectNameValidator.problem(it, emptyList()), it)
        }
    }

    @Test
    fun refusesATakenName() {
        assertEquals("web exists already: pick another name.", ProjectNameValidator.problem("web", listOf("web", "api")))
    }
}
