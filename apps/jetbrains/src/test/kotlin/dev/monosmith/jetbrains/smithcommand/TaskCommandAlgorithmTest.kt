package dev.monosmith.jetbrains.smithcommand

import kotlin.test.Test
import kotlin.test.assertEquals

class TaskCommandAlgorithmTest {
    @Test
    fun buildsTheRunCommand() {
        assertEquals(listOf("smith", "run", "test", "-p", "api"), TaskCommandAlgorithm.runTask("smith", "api", "test"))
    }

    @Test
    fun buildsTheProjectsCommand() {
        assertEquals(listOf("smith", "projects", "--json"), TaskCommandAlgorithm.listProjects("smith"))
    }

    @Test
    fun quotesArgumentsWithSpacesInTheDisplayedLine() {
        assertEquals("\"C:\\My Dir\\smith\" run test", TaskCommandAlgorithm.display(listOf("C:\\My Dir\\smith", "run", "test")))
    }
}
