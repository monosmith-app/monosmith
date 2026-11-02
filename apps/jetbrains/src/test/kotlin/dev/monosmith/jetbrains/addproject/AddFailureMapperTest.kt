package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class AddFailureMapperTest {
    private val failing =
        """
        Wrote 12 files
        smith: the new project desk was written but fails smith doctor
        The new project desk fails smith doctor:
          ✖ pkg-config does not find webkit2gtk-4.1
          ✖ the tauri CLI is missing
        Not committed: the checks fail.
        """.trimIndent()

    @Test
    fun readsAProjectThatStaysOnDiskAndFailsDoctor() {
        val failure = AddFailureMapper.doctorFailure(failing)!!
        assertEquals("desk", failure.name)
        assertEquals(12, failure.files)
        assertEquals(listOf("pkg-config does not find webkit2gtk-4.1", "the tauri CLI is missing"), failure.reasons)
        assertFalse(failure.committed)
    }

    @Test
    fun saysCreatedNeverRefusedAndThatNothingWasCommitted() {
        val text = AddFailureMapper.describe(AddFailureMapper.doctorFailure(failing)!!)
        assertTrue(text.startsWith("The project desk was created (12 files)"))
        assertTrue(text.contains("Nothing was committed"))
    }

    @Test
    fun anyOtherFailureIsNotAProjectOnDisk() {
        assertNull(AddFailureMapper.doctorFailure("smith: name taken"))
        assertNull(AddFailureMapper.missingTool("smith: name taken"))
    }

    @Test
    fun readsAMissingToolWithItsPageAndTheFirstCheck() {
        val output =
            """
            Creating node-app web in apps/web/:
            Wrote 9 files
            Verifying: smith run test -p web
            smith: npm is not installed or not on PATH. Install it from https://nodejs.org/en/download, then run it again
            """.trimIndent()
        val tool = AddFailureMapper.missingTool(output)!!
        assertEquals("Node.js", tool.title)
        assertEquals(9, tool.files)
        assertEquals("https://nodejs.org/en/download", tool.url)
        assertEquals("apps/web", tool.dir)
        assertEquals(listOf("run", "test", "-p", "web"), tool.check)
    }
}
