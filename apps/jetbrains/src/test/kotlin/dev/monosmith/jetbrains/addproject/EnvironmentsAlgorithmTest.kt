package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class EnvironmentsAlgorithmTest {
    @Test
    fun aViteWebAppIsBuiltOncePerEnvironment() {
        assertTrue(EnvironmentsAlgorithm.isWebApp(AddRequestContract("app", "web", "node", with = listOf("react-ts"))))
        assertTrue(EnvironmentsAlgorithm.isWebApp(AddRequestContract("app", "web", "node", stack = "vue")))
        assertFalse(EnvironmentsAlgorithm.isWebApp(AddRequestContract("lib", "web", "node", with = listOf("react"))))
        assertFalse(EnvironmentsAlgorithm.isWebApp(AddRequestContract("app", "api", "go", with = listOf("react"))))
        assertFalse(EnvironmentsAlgorithm.isWebApp(AddRequestContract("app", "api", "node", with = listOf("fastify"))))
    }

    @Test
    fun parsesNamesSeparatedByCommasSpacesOrSemicolons() {
        assertEquals(listOf("lcl", "dev", "uat"), EnvironmentsAlgorithm.parse("lcl, dev;uat "))
    }

    @Test
    fun refusesBadNamesWithTheFix() {
        assertNull(EnvironmentsAlgorithm.problem("lcl, dev, uat, prd"))
        assertNotNull(EnvironmentsAlgorithm.problem(""))
        assertTrue(EnvironmentsAlgorithm.problem("Dev")!!.contains("lowercase"))
        assertTrue(EnvironmentsAlgorithm.problem("production")!!.contains("Vite"))
        assertEquals("An environment is listed twice.", EnvironmentsAlgorithm.problem("dev dev"))
    }

    @Test
    fun theDefaultEnvironmentIsPrdElseTheLast() {
        assertEquals("prd", EnvironmentsAlgorithm.defaultOf(listOf("lcl", "prd", "dev")))
        assertEquals("uat", EnvironmentsAlgorithm.defaultOf(listOf("lcl", "uat")))
    }

    @Test
    fun givesTheFourFlagsOfSmithAdd() {
        assertEquals(
            listOf(
                "--environments",
                "lcl,prd",
                "--default-environment",
                "prd",
                "--environment-output",
                "project",
                "--env-files",
                "example",
            ),
            EnvironmentsAlgorithm.arguments(EnvironmentsContract(listOf("lcl", "prd"), "prd", "project", "example")),
        )
    }
}
