package dev.monosmith.jetbrains.installdependency

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class InstallDependencyUseCaseTest {
    private val fake = FakePortsHarness()
    private val projects =
        """
        [{"name":"web","dir":"apps/web","kind":"app","language":"node","tasks":["build"],"deps":[]},
         {"name":"utils","dir":"libs/utils","kind":"lib","language":"node","tasks":["build"],"deps":[]},
         {"name":"api","dir":"apps/api","kind":"app","language":"go","tasks":["build"],"deps":[]}]
        """.trimIndent()

    init {
        fake.answers["projects --json"] = fake.ok(projects)
    }

    @Test
    fun linksALibraryOfTheWorkspaceAfterTheCommandIsShown() {
        fake.ones.addAll(listOf("web", "workspace", "utils", "follow"))
        fake.plans.add(fake.yes())
        assertTrue(InstallDependencyUseCase.run(fake.ports))
        assertEquals(listOf("utils"), fake.rowsSeen[2].map { it.name })
        assertTrue(fake.plansSeen.single().startsWith("smith "))
        assertTrue("workspace:utils" in fake.calls.last().joinToString(" "))
        assertTrue(fake.shown.last())
        assertTrue(fake.infos.single().startsWith("The dependency is added. Not committed"))
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun installsPackagesFromTheRegistryAfterTheCommandIsShown() {
        fake.ones.addAll(listOf("web", "registry"))
        fake.texts.add("zod react@19")
        fake.plans.add(fake.yes())
        assertTrue(InstallDependencyUseCase.run(fake.ports))
        val line = fake.calls.last().joinToString(" ")
        assertTrue("zod" in line && "react@19" in line && "--registry" in line)
    }

    @Test
    fun theProjectAndTheSourceCanBeGivenSoTheyAreNotAsked() {
        fake.texts.add("zod")
        fake.plans.add(fake.yes())
        assertTrue(InstallDependencyUseCase.run(fake.ports, project = "web", only = "registry"))
        assertEquals(emptyList(), fake.rowsSeen)
    }

    @Test
    fun aGoProjectIsNotAskedForANodeRange() {
        fake.projectsWithALibrary()
        fake.ones.addAll(listOf("go-app", "workspace", "go-lib"))
        fake.plans.add(fake.yes())
        assertTrue(InstallDependencyUseCase.run(fake.ports))
        assertEquals(3, fake.rowsSeen.size)
    }

    @Test
    fun withNoLibraryToLinkItSaysWhyAndAsksNothingMore() {
        fake.ones.addAll(listOf("utils", "workspace"))
        assertFalse(InstallDependencyUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("No library of this workspace can be linked to utils."))
        assertTrue(fake.calls.none { "--registry" in it })
    }

    @Test
    fun cancelAtTheCommandChangesNothing() {
        fake.ones.addAll(listOf("web", "registry"))
        fake.texts.add("zod")
        assertFalse(InstallDependencyUseCase.run(fake.ports))
        assertEquals(1, fake.calls.size)
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aFailedInstallIsAWarningWithSmithsWords() {
        fake.ones.addAll(listOf("web", "registry"))
        fake.texts.add("zod")
        fake.plans.add(fake.yes())
        fake.answers["install"] = fake.fail("✖ the registry cannot be reached\n")
        assertFalse(InstallDependencyUseCase.run(fake.ports))
        assertEquals("Nothing was added. the registry cannot be reached", fake.warnings.single())
    }

    @Test
    fun aListOfProjectsItCannotGetIsAWarning() {
        fake.answers["projects --json"] = fake.fail("✖ smith.toml:3:9 is not valid\n")
        assertFalse(InstallDependencyUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("Nothing was added."))
    }

    private fun FakePortsHarness.projectsWithALibrary() {
        answers["projects --json"] =
            ok(
                """
                [{"name":"go-app","dir":"apps/go-app","kind":"app","language":"go","tasks":[],"deps":[]},
                 {"name":"go-lib","dir":"libs/go-lib","kind":"lib","language":"go","tasks":[],"deps":[]}]
                """.trimIndent(),
            )
    }
}
