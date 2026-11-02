package dev.monosmith.jetbrains.releaseflow

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ReleaseArgumentsAlgorithmTest {
    private val json =
        """{"results":[{"project":"cli","current":"1.2.3","next":"1.3.0","level":"minor"},{"project":"sdk","next":"0.1.0","level":"minor"},{"level":"patch"}],"unchanged":["docs","","kit"]}"""

    @Test
    fun readsThePackagesTheReleaseWouldPublish() {
        val packages = ReleaseArgumentsAlgorithm.planned(json)
        assertEquals(listOf("cli", "sdk"), packages.map { it.name })
        assertEquals("1.2.3", packages.first().current)
        assertEquals("", packages.last().current)
    }

    @Test
    fun readsThePackagesNoCommitReleases() {
        assertEquals(listOf("docs", "kit"), ReleaseArgumentsAlgorithm.unchanged(json))
        assertEquals(emptyList(), ReleaseArgumentsAlgorithm.unchanged("""{"results":[]}"""))
    }

    @Test
    fun aPlanIsADryRunAndThePublishAnswersYes() {
        assertEquals(listOf("release", "--dry-run"), ReleaseArgumentsAlgorithm.arguments(ReleaseChoicesContract(), publish = false))
        assertEquals(listOf("release", "--yes"), ReleaseArgumentsAlgorithm.arguments(ReleaseChoicesContract(), publish = true))
    }

    @Test
    fun holdsBumpsAndForcedReleasesBecomeFlags() {
        val choices = ReleaseChoicesContract(hold = listOf("a", "b"), bump = listOf("cli=major"), forced = listOf("docs=patch"))
        assertEquals(
            listOf("release", "--yes", "--hold", "a,b", "--bump", "cli=major", "--bump", "docs=patch", "--force"),
            ReleaseArgumentsAlgorithm.arguments(choices, publish = true),
        )
    }

    @Test
    fun noForceFlagWithoutAForcedPackage() {
        assertEquals(
            listOf("release", "--dry-run", "--bump", "cli=minor"),
            ReleaseArgumentsAlgorithm.arguments(ReleaseChoicesContract(bump = listOf("cli=minor")), publish = false),
        )
    }

    @Test
    fun saysWhatAForcedReleaseDoes() {
        assertEquals("", ReleaseArgumentsAlgorithm.forcedConsequences(emptyList()))
        val text = ReleaseArgumentsAlgorithm.forcedConsequences(listOf("docs=patch", "kit=minor"))
        assertTrue(text.contains("docs, kit: no commit releases it"))
        assertTrue(text.contains("HEAD"))
    }

    @Test
    fun theBumpQuestionKeepsTheBumpsFirstAndOffersOnlyPackagesAlreadyReleased() {
        val rows = ReleaseArgumentsAlgorithm.bumpRows(ReleaseArgumentsAlgorithm.planned(json), mapOf("cli" to "major"))
        assertEquals(listOf("", "cli"), rows.map { it.name })
        assertEquals("(default)", rows.first().description)
        assertTrue(rows.last().description.endsWith("major (changed)"))
    }

    @Test
    fun nothingIsHeldBackAtFirst() {
        assertTrue(ReleaseArgumentsAlgorithm.holdRows(ReleaseArgumentsAlgorithm.planned(json)).none { it.picked })
        assertEquals("(default)", ReleaseArgumentsAlgorithm.levelRows.first().description)
    }
}
