package dev.monosmith.jetbrains.workspace

import com.intellij.openapi.project.Project
import java.util.concurrent.CopyOnWriteArrayList

/** Tells the views that a flow changed the workspace, so they read smith again. */
object WorkspaceChangedStore {
    private val listeners = CopyOnWriteArrayList<(Project) -> Unit>()

    /** Registers [listener]; it runs on the thread of the flow, so it must hand work to the UI thread itself. */
    fun listen(listener: (Project) -> Unit) {
        listeners.add(listener)
    }

    /** Says that the workspace of [ide] changed. */
    fun changed(ide: Project) {
        listeners.forEach { it(ide) }
    }
}
