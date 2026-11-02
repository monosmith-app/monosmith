package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class AddRowsAlgorithmTest {
    private val react = OutcomeItemContract("react", "React", "a React app", "node", "framework", listOf("jsx"))
    private val electron = OutcomeItemContract("electron", "Electron", "desktop", "node", "shell", emptyList())
    private val web =
        OutcomeContract("web", "A website", "app", listOf("node"), "React", "People open it.", listOf("website"), listOf(react))
    private val kinds =
        listOf(KindOptionContract("node-app", "node", "Node app", "a node app"), KindOptionContract("go-lib", "go", "Go lib", "a go lib"))

    @Test
    fun theFirstQuestionListsTheOutcomesThenTheLongWayThenTheComposer() {
        val rows = AddRowsAlgorithm.outcomeRows(listOf(web))
        assertEquals(listOf("web", AddRowsAlgorithm.EVERY, AddRowsAlgorithm.COMPOSE), rows.map { it.name })
        assertTrue(rows.first().search.contains("jsx") && rows.first().search.contains("website"))
    }

    @Test
    fun theLanguageRowsUseTheNamesAPersonKnowsAndWhatTheLanguageOffers() {
        val outcome = web.copy(items = listOf(react, electron))
        val rows =
            AddRowsAlgorithm.languageRows(
                listOf(LanguageInfoContract("node", "TypeScript / JavaScript", "web and tools")),
                kinds,
                "app",
                listOf("node", "go"),
                outcome,
            )
        assertEquals(1, rows.size)
        assertEquals("TypeScript / JavaScript", rows.single().label)
        assertEquals("React, Electron", rows.single().description)
    }

    @Test
    fun theLongWayAsksTheRolesTheKindsHave() {
        assertEquals(listOf("app", "lib"), AddRowsAlgorithm.roleRows(kinds).map { it.name })
        assertTrue(AddRowsAlgorithm.roleRows(kinds).all { it.detail.isNotEmpty() })
    }

    @Test
    fun aWebsiteMayDecideLaterButAFunctionAppMayNot() {
        assertEquals(listOf(AddRowsAlgorithm.LATER, "react"), AddRowsAlgorithm.presetRows(listOf(react), "web").map { it.name })
        assertEquals(listOf("react"), AddRowsAlgorithm.presetRows(listOf(react), "function").map { it.name })
    }

    @Test
    fun followingTheSiblingsIsTheDefaultOfTheLikeQuestion() {
        val rows = AddRowsAlgorithm.likeRows(listOf("api", "web"))
        assertEquals(listOf("", "api", "web"), rows.map { it.name })
        assertTrue(rows.first().label.contains("(default)"))
    }

    @Test
    fun theBuildEnvironmentStartsOnPrd() {
        val rows = AddRowsAlgorithm.environmentRows(listOf("lcl", "dev", "prd"))
        assertEquals("prd", rows.first().name)
        assertEquals("(default)", rows.first().description)
        assertEquals("(default)", AddRowsAlgorithm.outputRows.first().description)
        assertEquals("(default)", AddRowsAlgorithm.envFileRows.first().description)
    }
}
