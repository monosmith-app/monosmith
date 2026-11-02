package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ReleasePreviewMapperTest {
    @Test
    fun readsThePlanWithTheReasonAndTheCommitsOfEachPackage() {
        val json =
            """{"channel":"next","results":[{"project":"cli","current":"1.2.3","next":"1.3.0-next.1","level":"minor",""" +
                """"reason":"feat: add x","commits":["abc feat: add x"]},""" +
                """{"project":"sdk","next":"0.1.0","forced":true}],"unchanged":["docs"]}"""
        val rows = ReleasePreviewMapper.rows(ReleasePreviewMapper.toPreview(json))
        assertEquals("Channel: next (its versions are prereleases)", rows.first().text)
        assertEquals("cli 1.2.3 → 1.3.0-next.1 · minor", rows[1].text)
        assertEquals(listOf("feat: add x", "abc feat: add x"), rows[1].children)
        assertEquals("sdk new → 0.1.0 · forced", rows[2].text)
        assertEquals("No commit releases: docs", rows.last().text)
    }

    @Test
    fun saysWhyNothingReleasesAndWhatToDoAboutIt() {
        val rows = ReleasePreviewMapper.rows(ReleasePreviewMapper.toPreview("""{"results":[],"unchanged":["docs","kit"]}"""))
        assertEquals("Channel: stable", rows.first().text)
        assertTrue(rows[1].text.startsWith("Nothing to release"))
        assertEquals(2, rows[1].children.size)
        assertTrue(rows[1].children.first().startsWith("docs: no commit releases it. Release… can release it anyway"))
    }

    @Test
    fun aPlanThatStoppedEarlyIsTheFirstRow() {
        val rows = ReleasePreviewMapper.rows(ReleasePreviewMapper.toPreview("""{"error":"no tags","results":[]}"""))
        assertEquals("The plan stopped: no tags", rows.first().text)
    }

    @Test
    fun skipsTheLinesPrintedBeforeTheJson() {
        val out =
            "plugin-sdk: would attach a.zip\nsmith: would publish npm x\n" +
                "{\n  \"results\": [{\"project\":\"cli\",\"next\":\"1.0.0\"}]\n}\n"
        assertEquals(
            "cli",
            ReleasePreviewMapper
                .toPreview(out)
                .results
                .single()
                .project,
        )
        assertEquals("{\"a\":1}", JsonFieldsMapper.fromFirstObject("{\"a\":1}"))
    }
}
