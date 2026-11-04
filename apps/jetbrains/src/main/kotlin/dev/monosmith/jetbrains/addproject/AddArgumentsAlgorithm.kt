package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig

/** The arguments of `smith add` for a request. */
object AddArgumentsAlgorithm {
    /**
     * @param mode `plan` adds `--dry-run` (smith changes nothing); `apply` adds `--yes` (the person agreed)
     * @param fresh overrides the request: give the stack's defaults and do not follow the siblings
     * @param commit with `apply`: `--commit`, so the change is one commit that `smith undo` reverts; without it smith commits nothing
     * @return the arguments after `smith`
     */
    fun arguments(
        request: AddRequestContract,
        apply: Boolean,
        fresh: Boolean = request.fresh,
        commit: Boolean = false,
    ): List<String> =
        buildList {
            addAll(listOf("add", request.role, request.name, SmithFlagsConfig.LANG_LONG, request.language))
            request.stack?.let { addAll(listOf("--stack", it)) }
            if (request.with.isNotEmpty()) addAll(listOf("--with", request.with.joinToString(",")))
            if (fresh) add("--fresh")
            if (!fresh) request.like?.let { addAll(listOf("--like", it)) }
            request.environments?.let { addAll(EnvironmentsAlgorithm.arguments(it)) }
            add(if (apply) "--yes" else "--dry-run")
            if (apply && commit) add("--commit")
        }

    /**
     * Whether the plan says the project follows its siblings (smith words it "follows <sibling>" in a plan
     * and "The N siblings agree on" in a run), so `--fresh` would change something.
     */
    fun followsSiblings(plan: String): Boolean = Regex("\\bfollows\\b|\\bagree on\\b", RegexOption.IGNORE_CASE).containsMatchIn(plan)
}
