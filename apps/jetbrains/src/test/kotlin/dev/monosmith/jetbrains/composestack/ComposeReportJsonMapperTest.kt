package dev.monosmith.jetbrains.composestack

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class ComposeReportJsonMapperTest {
    @Test
    fun readsTheCatalogueWithoutACheck() {
        val json =
            """{"layers":[{"id":"framework","title":"Framework","about":"how it is written"}],""" +
                """"items":[{"id":"node/react","kind":"choice","name":"react","axis":"framework","layer":"framework","many":false,"language":"node","default":true},{"id":"","name":"x"}]}"""
        val report = ComposeReportJsonMapper.toReport(json)
        assertEquals(1, report.items.size)
        assertEquals("react", report.items.single().title)
        assertTrue(report.items.single().isDefault)
        assertNull(report.check)
        assertNull(report.plan)
    }

    @Test
    fun readsAConflictAndAnOpenQuestion() {
        val json =
            """{"layers":[],"items":[],"check":{"language":"node","complete":false,"conflict":"A and B do not go together",""" +
                """"question":{"choice":"azure-functions","about":"the trigger","axis":"trigger","options":[{"name":"http","suggested":true},{"name":"timer"}]},"selection":["react"]}}"""
        val check = ComposeReportJsonMapper.toReport(json).check!!
        assertEquals("A and B do not go together", check.conflict)
        assertEquals("trigger", check.question!!.axis)
        assertEquals(listOf(true, false), check.question.options.map { it.suggested })
        assertEquals(listOf("react"), check.selection)
    }

    @Test
    fun readsThePlanWithItsCommandFilesAndTools() {
        val json =
            """{"layers":[],"items":[],"plan":{"command":"smith add app web","argv":["smith","add","app","web"],""" +
                """"files":[{"action":"create","path":"a","note":""}],""" +
                """"tools":[{"choice":"react","name":"node","title":"Node.js","state":"installed","command":""}],"notes":["n"],"recipeArgv":["smith","recipe","new"]}}"""
        val plan = ComposeReportJsonMapper.toReport(json).plan!!
        assertEquals(listOf("smith", "add", "app", "web"), plan.argv)
        assertEquals("installed", plan.tools.single().state)
        assertEquals(3, plan.recipeArgv.size)
    }

    @Test
    fun anEmptyPlanIsNoPlan() {
        assertNull(ComposeReportJsonMapper.toReport("""{"layers":[],"items":[],"plan":{}}""").plan)
    }
}
