package dev.monosmith.jetbrains.adoptfolder

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_SUGGESTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ANOTHER_FOLDER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_A_FOLDER_THAT_STARTS_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_IS_NOT_A_FOLDER_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_IT_IS_THE_REPOSITORY_ROOT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_IT_LEAVES_THE_REPOSITORY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_IT_MUST_BE_RELATIVE_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_MUST_BE_LOWERCASE_LETTERS_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NO_LEAVE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_THE_PROJECT_STAYS_AT_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_USE_LETTERS_DIGITS_DOT_HYPHEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/**
 * The one project a plain repository keeps at its root, where smith cannot manage it: it has to move
 * into a folder of its own.
 *
 * @property app true when it looks like a program, which makes an app folder the suggestion
 */
data class RootProjectContract(
    val manifest: String,
    val ecosystem: String,
    val name: String,
    val app: Boolean,
)

/**
 * What `smith adopt --json` found.
 *
 * @property blockers what stops the adoption, each a problem and its fix
 * @property providers the CI providers the repository already has
 */
data class AdoptReportContract(
    val name: String,
    val blockers: List<String>,
    val providers: List<String>,
    val projects: Int,
    val rootProject: RootProjectContract?,
)

/** One answer of "move it into a folder of its own?". [dir] is empty for another folder (typed) and for none. */
data class MoveChoiceContract(
    val key: String,
    val label: String,
    val dir: String,
    val suggested: Boolean,
)

/** Reads the adoption report and words the questions of `smith adopt --init`, the same ones a terminal asks. */
object AdoptReportMapper {
    /** The `--move` value that keeps the project where it is. */
    const val LEAVE = "none"

    private val namePattern = Regex("^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$")
    private val segmentPattern = Regex("^[A-Z0-9][\\w.-]*$", RegexOption.IGNORE_CASE)

    /**
     * @param json what `smith adopt --dry-run --json` printed
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toReport(json: String): AdoptReportContract {
        val root = JsonFieldsMapper.parseObject(json)
        return AdoptReportContract(
            name = JsonFieldsMapper.text(root, "name"),
            blockers =
                JsonFieldsMapper.records(root, "blockers").map {
                    "${JsonFieldsMapper.text(it, "message")}. ${JsonFieldsMapper.text(it, "step")}"
                },
            providers = JsonFieldsMapper.words(root, "providers"),
            projects = JsonFieldsMapper.records(root, "projects").size,
            rootProject =
                JsonFieldsMapper.record(root, "rootProject")?.let {
                    RootProjectContract(
                        JsonFieldsMapper.text(it, "manifest"),
                        JsonFieldsMapper.text(it, "ecosystem"),
                        JsonFieldsMapper.text(it, "name"),
                        JsonFieldsMapper.flag(it, "app"),
                    )
                },
        )
    }

    /** The five answers, in the order smith numbers them. The suggested one is an app folder for a program, else a package folder. */
    fun moveChoices(project: RootProjectContract): List<MoveChoiceContract> {
        val suggested = if (project.app) "apps" else "packages"
        return listOf(
            "apps",
            "packages",
            "libs",
        ).map { MoveChoiceContract(it, "$it/${project.name}", "$it/${project.name}", it == suggested) } +
            listOf(
                MoveChoiceContract("other", say(JETBRAINS_ADOPTFOLDER_ANOTHER_FOLDER), "", false),
                MoveChoiceContract("none", say(JETBRAINS_ADOPTFOLDER_NO_LEAVE_IT), "", false),
            )
    }

    /** The lines of the move question; the suggested answer is first and marked. */
    fun moveRows(project: RootProjectContract): List<ChoiceRowContract> =
        moveChoices(project).sortedByDescending { it.suggested }.map {
            ChoiceRowContract(
                it.key,
                it.label,
                if (it.suggested) say(JETBRAINS_ADDCHOICES_SUGGESTED) else "",
                if (it.key ==
                    "none"
                ) {
                    say(JETBRAINS_ADOPTFOLDER_THE_PROJECT_STAYS_AT_THE)
                } else {
                    ""
                },
            )
        }

    /** What is wrong with a project name, or null. */
    fun nameProblem(name: String): String? =
        if (name.length <= 50 && namePattern.matches(name)) {
            null
        } else {
            say(JETBRAINS_ADOPTFOLDER_MUST_BE_LOWERCASE_LETTERS_AND, name.toString())
        }

    /** Checks a folder typed relative to the repository root: the cleaned folder, or a sentence saying why it cannot be used. */
    fun folderOrProblem(folder: String): Pair<String?, String?> {
        val raw = folder.trim().replace('\\', '/')

        fun bad(why: String) = null to say(JETBRAINS_ADOPTFOLDER_IS_NOT_A_FOLDER_SMITH, folder.toString(), why.toString())
        if (raw in listOf("", ".", "./")) return bad(say(JETBRAINS_ADOPTFOLDER_IT_IS_THE_REPOSITORY_ROOT))
        if (raw.startsWith("/") || Regex("^[A-Za-z]:").containsMatchIn(raw)) return bad(say(JETBRAINS_ADOPTFOLDER_IT_MUST_BE_RELATIVE_TO))
        val segments = raw.split("/").filter { it.isNotEmpty() && it != "." }
        for (segment in segments) {
            if (segment == "..") return bad(say(JETBRAINS_ADOPTFOLDER_IT_LEAVES_THE_REPOSITORY))
            if (segment.startsWith(".")) return bad(say(JETBRAINS_ADOPTFOLDER_A_FOLDER_THAT_STARTS_WITH))
            if (!segmentPattern.matches(segment)) return bad(say(JETBRAINS_ADOPTFOLDER_USE_LETTERS_DIGITS_DOT_HYPHEN))
        }
        return segments.joinToString("/") to null
    }

    /** The arguments of `smith adopt --init` for the answers; every open question has its flag, so smith asks nothing. */
    fun initArguments(
        move: String?,
        gitattributes: String,
        agent: String,
        ci: List<String>,
        ciExisting: String?,
        apply: Boolean,
    ): List<String> =
        buildList {
            addAll(listOf("adopt", "--init"))
            if (move != null) addAll(listOf("--move", move))
            addAll(listOf("--gitattributes", gitattributes, "--agent", agent, "--ci", if (ci.isEmpty()) "none" else ci.joinToString(",")))
            if (ciExisting != null) addAll(listOf("--ci-existing", ciExisting))
            add(if (apply) "--yes" else "--dry-run")
        }
}
