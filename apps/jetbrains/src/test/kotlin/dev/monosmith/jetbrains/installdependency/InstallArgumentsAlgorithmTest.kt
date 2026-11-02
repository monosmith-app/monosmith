package dev.monosmith.jetbrains.installdependency

import kotlin.test.Test
import kotlin.test.assertEquals

class InstallArgumentsAlgorithmTest {
    @Test
    fun aLibraryIsNamedWithTheWorkspacePrefixSoSmithNeverAsks() {
        assertEquals(listOf("install", "-w", "web", "workspace:ui"), InstallArgumentsAlgorithm.library("web", "ui", null))
        assertEquals(listOf("install", "-w", "web", "workspace:ui", "--range", "*"), InstallArgumentsAlgorithm.library("web", "ui", "*"))
    }

    @Test
    fun aRegistryPackageCarriesRegistry() {
        assertEquals(
            listOf("install", "-w", "web", "--registry", "zod", "react@19"),
            InstallArgumentsAlgorithm.registry("web", listOf("zod", "react@19"), dev = false),
        )
        assertEquals(
            listOf("install", "-w", "web", "-D", "--registry", "vitest"),
            InstallArgumentsAlgorithm.registry("web", listOf("vitest"), dev = true),
        )
    }

    @Test
    fun unlinkNamesTheLibraryAndAnswersYes() {
        assertEquals(listOf("uninstall", "-w", "web", "workspace:ui", "--yes"), InstallArgumentsAlgorithm.unlink("web", "ui"))
    }

    @Test
    fun splitsPackageNamesOnSpacesAndCommas() {
        assertEquals(listOf("zod", "react@19", "x"), InstallArgumentsAlgorithm.parsePackages("zod, react@19  x"))
        assertEquals(emptyList(), InstallArgumentsAlgorithm.parsePackages("  "))
    }
}
