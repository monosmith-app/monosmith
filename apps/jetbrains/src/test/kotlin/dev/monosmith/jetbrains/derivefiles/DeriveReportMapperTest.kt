package dev.monosmith.jetbrains.derivefiles

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class DeriveReportMapperTest {
    private val out =
        """
        2 derived files need action:
          Drift      packages/web/tsconfig.json  (tsconfig-references)  edited by hand
          Stale      smith.lock  (lock)  inputs changed
          Odd        x  (y)  z
        5 derived files in all
        """.trimIndent()

    @Test
    fun readsTheFilesThatNeedActionAndIgnoresALineItDoesNotKnow() {
        val files = DeriveReportMapper.files(out)
        assertEquals(listOf("packages/web/tsconfig.json", "smith.lock"), files.map { it.path })
        assertEquals("drift", files.first().status)
        assertEquals("tsconfig-references", files.first().renderer)
        assertEquals("edited by hand", files.first().detail)
    }

    @Test
    fun aDriftedDivergedOrBrokenFileIsHeldBack() {
        assertTrue(DeriveReportMapper.isHeldBack("drift"))
        assertTrue(DeriveReportMapper.isHeldBack("broken"))
        assertFalse(DeriveReportMapper.isHeldBack("stale"))
    }

    @Test
    fun theFileGoesInFlagEqualsFileFormSoADashNameIsNeverAnOption() {
        assertEquals(listOf("derive", "--check"), DeriveReportMapper.arguments("check"))
        assertEquals(listOf("derive", "--dry-run"), DeriveReportMapper.arguments("preview"))
        assertEquals(listOf("derive", "--yes"), DeriveReportMapper.arguments("apply"))
        assertEquals(listOf("derive", "--explain=-a.json"), DeriveReportMapper.arguments("explain", "-a.json"))
        assertEquals(listOf("derive", "--reclaim=a.json", "--yes"), DeriveReportMapper.arguments("reclaim", "a.json"))
        assertEquals(listOf("derive", "--replace=a.json", "--yes"), DeriveReportMapper.arguments("replace", "a.json"))
    }
}
