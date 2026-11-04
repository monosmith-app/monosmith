package dev.monosmith.jetbrains.composestack

import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig

/** The pure steps of composing a stack: the arguments, the words of a plan, and the radio behaviour of the picks. */
object ComposeStackAlgorithm {
    /**
     * The arguments of `smith compose --json` for the picks so far; no `--with` when nothing is ticked.
     *
     * @param kind the kind to plan for, with [name]
     */
    fun arguments(
        language: String?,
        picks: List<String>,
        kind: String? = null,
        name: String? = null,
    ): List<String> =
        buildList {
            addAll(listOf("compose", "--json"))
            if (!language.isNullOrEmpty()) addAll(listOf(SmithFlagsConfig.LANG_LONG, language))
            if (picks.isNotEmpty()) addAll(listOf("--with", picks.joinToString(",")))
            kind?.let { addAll(listOf("--kind", it)) }
            name?.let { addAll(listOf("--name", it)) }
        }

    /** One line for a tool of the plan: the name, whether it is there and which pick needs it, and the command that installs it when it is not. */
    fun describeTool(tool: ComposeToolContract): String {
        val state = tool.state.ifEmpty { "not checked" }
        val line = "${tool.title} ($state${if (tool.choice.isEmpty()) "" else ", for ${tool.choice}"})"
        return if (tool.state == "missing" || tool.state == "outdated") "$line: ${tool.command}" else line
    }

    /** The tools of a plan that are not there or too old, in the order of the plan. */
    fun toolsToInstall(plan: ComposePlanContract): List<ComposeToolContract> =
        plan.tools.filter {
            it.state == "missing" ||
                it.state == "outdated"
        }

    /** Words the plan for the confirmation: the command, the files, the tools and the notes. */
    fun describePlan(plan: ComposePlanContract): String {
        val lines = mutableListOf("Run: ${plan.command}", "")
        if (plan.files.isNotEmpty()) {
            lines.add("It writes:")
            plan.files.forEach { lines.add("  ${it.action} ${it.path}${if (it.note.isEmpty()) "" else " (${it.note})"}") }
            lines.add("")
        }
        if (plan.tools.isNotEmpty()) {
            lines.add("It needs these tools on this computer:")
            plan.tools.forEach { lines.add("  ${describeTool(it)}") }
            lines.add("")
        }
        lines.addAll(plan.notes)
        return lines.joinToString("\n").trim()
    }

    /**
     * The lines a list shows for an item: title with its mark, where it is for, what it is. [detail] holds the
     * aliases, so typing "desktop" finds Electron.
     */
    fun labels(item: ComposeItemContract): Triple<String, String, String> {
        val where = if (item.language.isEmpty() || item.kind == "language") "" else "${item.language} · "
        // Electron sits on the shell layer but answers the framework question; say so.
        val slot =
            if (item.kind == "choice" && item.axis.isNotEmpty() && item.axis != item.layer &&
                !item.many
            ) {
                " (the ${item.axis} choice)"
            } else {
                ""
            }
        val mark =
            if (item.isDefault) {
                " (default)"
            } else if (item.suggested) {
                " (suggested)"
            } else {
                ""
            }
        val problem = if (item.problem.isEmpty()) "" else " Cannot be used: ${item.problem}."
        val route = if (item.route.isEmpty()) "" else " Not part of one project: turn it on with ${item.route}."
        return Triple(
            "${item.title}$mark",
            "$where${item.layer}$slot",
            "${item.description}$problem$route ${item.aliases.joinToString(" ")}".trim(),
        )
    }

    /** Whether a typed search finds an item: every word must start a word of the title, name, alias or description. */
    fun matches(
        item: ComposeItemContract,
        query: String,
    ): Boolean {
        val words = query.lowercase().split(Regex("[^\\p{L}\\p{N}#+.]+")).filter { it.isNotEmpty() }
        if (words.isEmpty()) return true
        val own =
            "${item.title} ${item.name} ${item.aliases.joinToString(
                " ",
            )} ${item.description} ${item.layer} ${item.language}".lowercase().split(Regex("[^\\p{L}\\p{N}#+.]+"))
        return words.all { word -> own.any { it.startsWith(word) } }
    }

    /** The titles of the ticked items, in catalogue order, to show what is chosen when a search hides the rows. */
    fun tickedTitles(
        items: List<ComposeItemContract>,
        ids: Collection<String>,
    ): String = items.filter { it.id in ids }.joinToString(", ") { it.title }

    /** The stack in words, to name it where a question about it is asked: "React + Electron (node)". */
    fun stackSummary(
        items: List<ComposeItemContract>,
        language: String,
        picks: List<String>,
    ): String {
        val titles =
            picks.map { name ->
                items.firstOrNull { it.kind == "choice" && it.name == name && it.language == language }?.title
                    ?: name
            }
        val lang = items.firstOrNull { it.kind == "language" && it.name == language }?.title ?: language
        return "${titles.joinToString(" + ")}${if (lang.isEmpty()) "" else " ($lang)"}"
    }

    /**
     * Keeps one pick per single-choice axis, like radio buttons: a choice ticked in this round takes the
     * place of a choice ticked before on the same axis and language, unless both allow several. The
     * catalogue decides the axes; no name is known here.
     *
     * @param previous the ids ticked before this round
     * @param chosen the ids ticked now
     * @return the ids to keep, and a one-line note naming what was replaced (empty when nothing was)
     */
    fun settleRivals(
        items: List<ComposeItemContract>,
        previous: List<String>,
        chosen: List<String>,
    ): Pair<List<String>, String> {
        val byId = items.associateBy { it.id }
        val fresh = chosen.filter { it !in previous }
        val dropped = mutableSetOf<String>()
        val notes = mutableListOf<String>()
        for (id in fresh) {
            val pick = byId[id] ?: continue
            val isLanguage = pick.kind == "language"
            if (!isLanguage && (pick.kind != "choice" || pick.axis.isEmpty())) continue
            val later = fresh.drop(fresh.indexOf(id) + 1)
            val rivals =
                chosen.filter { other ->
                    val item = byId[other]
                    when {
                        other == id || other in later || item == null || other in dropped -> {
                            false
                        }

                        isLanguage -> {
                            item.kind == "language"
                        }

                        else -> {
                            item.kind == "choice" && item.axis == pick.axis && item.language == pick.language &&
                                !(item.many && pick.many)
                        }
                    }
                }
            dropped.addAll(rivals)
            if (rivals.isNotEmpty()) {
                notes.add(
                    "${pick.title} replaces ${rivals.joinToString(
                        " and ",
                    ) { byId[it]?.title ?: it }}: a project has one ${if (isLanguage) "language" else pick.axis}.",
                )
            }
        }
        return chosen.filter { it !in dropped } to notes.joinToString(" ")
    }
}
