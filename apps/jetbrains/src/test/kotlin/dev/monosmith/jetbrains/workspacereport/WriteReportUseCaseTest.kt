package dev.monosmith.jetbrains.workspacereport

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import java.nio.file.Path
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class WriteReportUseCaseTest {
    private val fake = FakePortsHarness()
    private val file = WriteReportUseCase.reportFile(Path.of("tmp"), "my repo!", 7)

    @Test
    fun theFileNameIsSafeAndNewEachTime() {
        assertEquals("monosmith-report-my-repo--7.html", file.fileName.toString())
        assertEquals("monosmith-report-workspace-1.html", WriteReportUseCase.reportFile(Path.of("tmp"), "", 1).fileName.toString())
    }

    @Test
    fun writesThePageAndOpensIt() {
        var opened: Path? = null
        assertTrue(WriteReportUseCase.run(fake.ports, file) { opened = it })
        assertEquals(listOf("report", "-o", file.toString()), fake.calls.single())
        assertEquals(file, opened)
    }

    @Test
    fun aFailureIsAWarningAndOpensNothing() {
        fake.answers["report"] = fake.fail("✖ the workspace has no smith.toml")
        var opened = false
        assertFalse(WriteReportUseCase.run(fake.ports, file) { opened = true })
        assertFalse(opened)
        assertTrue(fake.warnings.single().startsWith("The report was not written."))
    }
}
