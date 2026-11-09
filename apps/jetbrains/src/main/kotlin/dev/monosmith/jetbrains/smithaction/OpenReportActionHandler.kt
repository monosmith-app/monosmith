package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacereport.WriteReportUseCase

/** Writes the report page of the workspace and opens it. */
class OpenReportActionHandler : SmithActionHandler({ ide, root -> WriteReportUseCase.start(ide, root) })
