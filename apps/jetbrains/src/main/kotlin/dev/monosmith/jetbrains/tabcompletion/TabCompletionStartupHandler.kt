package dev.monosmith.jetbrains.tabcompletion

import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.ProjectActivity
import java.nio.file.Path

/**
 * Offers tab completion once after a project opens, in a notification that writes nothing. It asks smith
 * (`smith completion status --json`) first, so a person who has it, or who said no on the command line,
 * is not asked.
 */
class TabCompletionStartupHandler :
    ProjectActivity,
    DumbAware {
    override suspend fun execute(project: Project) {
        val folder = project.basePath?.let(Path::of) ?: return
        OfferTabCompletionUseCase.offer(TabCompletionIdeClient(project, folder))
    }
}
