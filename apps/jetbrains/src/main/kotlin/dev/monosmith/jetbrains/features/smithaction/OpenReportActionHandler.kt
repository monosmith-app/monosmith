package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacereport.WriteReportUseCase

/** Writes the report page of the workspace and opens it. */
class OpenReportActionHandler : SmithActionHandler({ ide, root -> WriteReportUseCase.start(ide, root) })
