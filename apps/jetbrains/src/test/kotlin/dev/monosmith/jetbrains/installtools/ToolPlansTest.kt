package dev.monosmith.jetbrains.installtools

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ToolPlansTest {
    private val json =
        """
        [{"name":"node","title":"Node.js","why":"runs node projects","consequence":"node projects cannot run","download":"https://nodejs.org","state":"missing",
          "default":"install:winget",
          "options":[
            {"kind":"page","value":"page","label":"Open the page","hint":"download it yourself"},
            {"kind":"install","value":"install:winget","manager":"winget","command":["winget","install","OpenJS.NodeJS"],"admin":true,"label":"winget","hint":"installs it","note":"asks for approval"},
            {"kind":"skip","value":"skip","label":"Skip","hint":"leave it"}]}]
        """.trimIndent()

    @Test
    fun readsThePlanAndItsOptions() {
        val plan = ToolPlansJsonMapper.toPlans(json).single()
        assertEquals("Node.js", plan.title)
        assertEquals("install:winget", plan.default)
        assertEquals(listOf("page", "install", "skip"), plan.options.map { it.kind })
        assertTrue(plan.options[1].admin)
        assertEquals(listOf("winget", "install", "OpenJS.NodeJS"), plan.options[1].command)
    }

    @Test
    fun nothingMissingIsNullOrAnEmptyList() {
        assertEquals(emptyList(), ToolPlansJsonMapper.toPlans("null"))
        assertEquals(emptyList(), ToolPlansJsonMapper.toPlans("[]"))
    }

    @Test
    fun theMarkedDefaultComesFirstAndEachRowShowsItsCommandAndWhetherItNeedsAdmin() {
        val plan = ToolPlansJsonMapper.toPlans(json).single()
        val rows = ToolRowsAlgorithm.wayRows(plan)
        assertEquals(listOf("install:winget", "page", "skip"), rows.map { it.name })
        assertEquals("(default)", rows.first().description)
        assertTrue(rows.first().detail.startsWith("Runs: winget install OpenJS.NodeJS."))
        assertTrue(rows.first().detail.contains("administrator rights"))
    }

    @Test
    fun theAdminQuestionStartsOnSkipSoNothingRunsByAccident() {
        val plan = ToolPlansJsonMapper.toPlans(json).single()
        val rows = ToolRowsAlgorithm.adminRows(plan, plan.options[1])
        assertEquals(listOf(ToolRowsAlgorithm.SKIP, ToolRowsAlgorithm.MYSELF, ToolRowsAlgorithm.RUN_IT), rows.map { it.name })
        assertEquals("(default)", rows.first().description)
    }

    @Test
    fun listsEveryInstallCommandThenThePage() {
        assertEquals(
            listOf("winget install OpenJS.NodeJS", "https://nodejs.org"),
            ToolRowsAlgorithm.commands(ToolPlansJsonMapper.toPlans(json).single()),
        )
    }
}
