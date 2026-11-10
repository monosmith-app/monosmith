package dev.monosmith.jetbrains.installdependency

import dev.monosmith.jetbrains.workspace.SmithProjectContract

/** The rule `smith install` applies when a project depends on a library of the workspace. */
object LinkableLibrariesPolicy {
    /**
     * Says whether a project may depend on a library, the way `smith install` decides: not itself, not an
     * app, the same language, and no cycle (the library must not already depend on the project, directly or not).
     */
    fun mayLink(
        nodes: List<SmithProjectContract>,
        from: String,
        library: String,
    ): Boolean {
        val byName = nodes.associateBy { it.name }
        val a = byName[from] ?: return false
        val b = byName[library] ?: return false
        if (a.name == b.name || b.kind == "app" || a.language != b.language) return false
        val seen = mutableSetOf(b.name)
        val queue = ArrayDeque(listOf(b.name))
        while (queue.isNotEmpty()) {
            val next = queue.removeFirst()
            if (next == a.name) return false
            for (dep in byName[next]?.deps.orEmpty()) if (seen.add(dep)) queue.addLast(dep)
        }
        return true
    }

    /** The libraries a project may depend on, in the order given. */
    fun linkable(
        nodes: List<SmithProjectContract>,
        from: String,
    ): List<SmithProjectContract> = nodes.filter { mayLink(nodes, from, it.name) }
}
