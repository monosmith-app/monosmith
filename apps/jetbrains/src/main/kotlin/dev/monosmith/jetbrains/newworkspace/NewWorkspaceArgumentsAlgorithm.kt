package dev.monosmith.jetbrains.newworkspace

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithcommand.SmithFlags

/**
 * What the person decides about a new workspace: the questions of `smith new`, one field each.
 *
 * @property ci where the tests run on every push
 * @property language the main language, or `later`
 * @property nodeStack the tests and lint of Node projects; used only for `node`
 * @property architecture the code layout smith checks, or `none`
 * @property agent write the MonoSmith agent files
 * @property devcontainer write the dev container
 * @property commit make the first commit of smith's own files
 */
data class NewWorkspaceChoicesContract(
    val name: String,
    val ci: List<String> = listOf("github"),
    val language: String = "later",
    val nodeStack: String = "node-jest",
    val architecture: String = "vertical-slices",
    val agent: Boolean = true,
    val devcontainer: Boolean = true,
    val commit: Boolean = true,
)

/** The command line of `smith new` and the summary of the choices. The lists of answers come from `smith new --questions --json`. */
object NewWorkspaceArgumentsAlgorithm {
    /** The defaults the questions say: what "Start with the defaults" chooses. */
    fun defaults(
        name: String,
        questions: List<NewQuestionContract>,
    ): NewWorkspaceChoicesContract {
        fun of(key: String) = questions.firstOrNull { it.key == key }?.default
        return NewWorkspaceChoicesContract(
            name = name,
            ci = of("ci")?.split(",")?.filter { it.isNotEmpty() } ?: listOf("github"),
            language = of("language") ?: "later",
            nodeStack = of("node-stack") ?: "node-jest",
            architecture = of("architecture") ?: "vertical-slices",
            agent = (of("agent") ?: "yes") == "yes",
            devcontainer = (of("devcontainer") ?: "yes") == "yes",
        )
    }

    /** The rows of a yes or no question for the first commit, which `smith new` has as a flag and not as a question; yes is first when [defaultYes]. */
    fun yesNoRows(
        yes: String,
        no: String,
        defaultYes: Boolean,
    ): List<ChoiceRowContract> {
        val rows =
            listOf(
                ChoiceRowContract("yes", "Yes", if (defaultYes) "(default)" else "", yes),
                ChoiceRowContract("no", "No", if (defaultYes) "" else "(default)", no),
            )
        return if (defaultYes) rows else rows.reversed()
    }

    /**
     * The arguments of `smith new` for the choices. Every question has a flag, so smith asks nothing more
     * and decides nothing in silence. `--here` creates the workspace in the (empty) folder the IDE has open.
     *
     * @param preview add `--dry-run`, which lists the files and writes nothing
     */
    fun arguments(
        choices: NewWorkspaceChoicesContract,
        preview: Boolean,
    ): List<String> =
        buildList {
            addAll(listOf("new", choices.name, "--here", "--yes", "--ci", choices.ci.joinToString(","), SmithFlags.LANG_LONG, choices.language))
            if (choices.language == "node") addAll(listOf("--default-stack", "node=${choices.nodeStack}"))
            addAll(listOf("--architecture", choices.architecture))
            add(if (choices.agent) "--agent" else "--no-agent")
            add(if (choices.devcontainer) "--devcontainer" else "--no-devcontainer")
            if (!choices.commit) add("--no-commit")
            if (preview) add("--dry-run")
        }

    /** What the choices mean, one line each, for the summary shown before smith writes anything. */
    fun describe(
        choices: NewWorkspaceChoicesContract,
        folderName: String,
        questions: List<NewQuestionContract>,
    ): List<String> {
        fun q(key: String) = questions.firstOrNull { it.key == key }
        return listOf(
            "Name: ${choices.name}${if (choices.name == folderName) " (the name of this folder)" else ""}",
            "CI: ${choices.ci.joinToString(", ") { NewQuestionsMapper.labelOf(q("ci"), it) }}",
            "Main language: ${NewQuestionsMapper.labelOf(q("language"), choices.language)}",
        ) +
            (
                if (choices.language ==
                    "node"
                ) {
                    listOf("Tests and lint for Node projects: ${NewQuestionsMapper.labelOf(q("node-stack"), choices.nodeStack)}")
                } else {
                    emptyList()
                }
            ) +
            listOf(
                "Code layout smith checks: ${choices.architecture}",
                "Agent files for Claude Code: ${if (choices.agent) "yes" else "no"}",
                "Dev container: ${if (choices.devcontainer) "yes" else "no"}",
                "First commit of smith's own files: ${if (choices.commit) "yes, smith makes it" else "no, the files stay uncommitted"}",
            )
    }
}
