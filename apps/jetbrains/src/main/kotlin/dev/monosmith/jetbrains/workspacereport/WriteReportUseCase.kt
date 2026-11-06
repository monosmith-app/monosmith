package dev.monosmith.jetbrains.workspacereport

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACEREPORT_MONOSMITH_REPORT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACEREPORT_THE_REPORT_WAS_NOT_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Asks smith for the report page of the workspace (`smith report -o <file>`): the map of the projects,
 * health, derived-file drift and choices. The file goes in the temp folder with a name of its own each
 * time, so the workspace stays clean and smith never refuses to replace a file.
 */
object WriteReportUseCase {
    /** The file of a report: a safe name from the workspace folder and a time stamp. */
    fun reportFile(
        tempDir: Path,
        folderName: String,
        stamp: Long,
    ): Path = tempDir.resolve("monosmith-report-${folderName.replace(Regex("[^\\w.-]+"), "-").ifEmpty { "workspace" }}-$stamp.html")

    /** Starts the flow in a background task and opens the page in the browser. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACEREPORT_MONOSMITH_REPORT)) {
        val file =
            reportFile(Path.of(System.getProperty("java.io.tmpdir")), root.fileName?.toString().orEmpty(), System.currentTimeMillis())
        run(IdePortsClient.of(ide, root), file) { BrowserUtil.browse(it.toUri()) }
    }

    /**
     * Writes the report to [file] and opens it.
     *
     * @param open shows the page
     * @return true when the page was written
     */
    fun run(
        ports: FlowPortsContract,
        file: Path,
        open: (Path) -> Unit,
    ): Boolean {
        val done = ports.smith.run(listOf("report", SmithFlagsConfig.OUTPUT, file.toString()))
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_WORKSPACEREPORT_THE_REPORT_WAS_NOT_WRITTEN, FailureTextMapper.describe(done).toString()))
            return false
        }
        open(file)
        return true
    }
}
