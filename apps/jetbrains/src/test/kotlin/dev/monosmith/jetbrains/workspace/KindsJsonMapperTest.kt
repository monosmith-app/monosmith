package dev.monosmith.jetbrains.workspace

import kotlin.test.Test
import kotlin.test.assertEquals

class KindsJsonMapperTest {
    @Test
    fun readsTheKindsAndTheirRole() {
        val kinds =
            KindsJsonMapper.toKinds(
                """[{"kind":"node-app","language":"node","label":"Node app"},{"kind":"go-lib","language":"go","label":"Go lib"}]""",
            )
        assertEquals(listOf("node-app", "go-lib"), kinds.map { it.kind })
        assertEquals(listOf("app", "lib"), kinds.map { it.role() })
        assertEquals("Go lib", kinds[1].toString())
    }
}
