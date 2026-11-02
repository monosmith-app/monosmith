package dev.monosmith.jetbrains.dockerhub

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class DockerDescriptionUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun namesTheRepositoriesFirstAndCallsDockerHubOnlyAfterAYes() {
        fake.answers["docker description --dry-run"] = fake.ok("would set acme/web\n")
        fake.plans.add(fake.yes())
        assertTrue(DockerDescriptionUseCase.run(fake.ports))
        assertEquals("would set acme/web", fake.plansSeen.single())
        assertEquals(listOf("docker", "description"), fake.calls.last())
    }

    @Test
    fun cancelCallsNothing() {
        assertFalse(DockerDescriptionUseCase.run(fake.ports))
        assertEquals(listOf(listOf("docker", "description", "--dry-run")), fake.calls)
    }
}
