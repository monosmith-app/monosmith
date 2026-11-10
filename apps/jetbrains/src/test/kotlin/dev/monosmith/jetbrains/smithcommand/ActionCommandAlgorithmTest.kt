package dev.monosmith.jetbrains.smithcommand

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse

class ActionCommandAlgorithmTest {
    @Test
    fun runsDoctorWithOrWithoutFix() {
        assertEquals(listOf("smith", "doctor"), ActionCommandAlgorithm.doctor("smith", fix = false))
        assertEquals(listOf("smith", "doctor", "--fix"), ActionCommandAlgorithm.doctor("smith", fix = true))
    }

    @Test
    fun addsWithoutCommittingUnlessAsked() {
        val plain = ActionCommandAlgorithm.add("smith", "app", "web", "node", commit = false)
        assertEquals(listOf("smith", "add", "app", "web", "--lang", "node", "--yes"), plain)
        assertFalse("--commit" in plain)
        assertEquals(
            listOf("smith", "add", "lib", "core", "--lang", "go", "--yes", "--commit"),
            ActionCommandAlgorithm.add("smith", "lib", "core", "go", commit = true),
        )
    }

    @Test
    fun buildsTheOtherCommands() {
        assertEquals(listOf("smith", "kinds", "--json"), ActionCommandAlgorithm.listKinds("smith"))
        assertEquals(listOf("smith", "undo", "--yes"), ActionCommandAlgorithm.undo("smith"))
        assertEquals(listOf("smith", "why", "apps/web/package.json"), ActionCommandAlgorithm.why("smith", "apps/web/package.json"))
        assertEquals(listOf("smith", "release", "--dry-run"), ActionCommandAlgorithm.releasePreview("smith"))
    }
}
