package dev.monosmith.jetbrains.adoptfolder

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.newworkspace.NewQuestionsMapper
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Brings a repository that has no smith.toml into MonoSmith (`smith adopt --init`). smith reports what it
 * found (`--dry-run`); the person then answers the four questions a terminal asks (line endings, agent
 * files, CI providers, a pipeline of their own) and, for a project at the repository root, where to move
 * it; the plan with diffs comes next, and only then smith writes. Every answer is a flag, so nothing is
 * asked twice; nothing is committed unless the box is ticked.
 */
object AdoptFolderUseCase {
    private const val TITLE = "MonoSmith: bring this folder in"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @return true when smith adopted the folder
     */
    fun run(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val found = ports.smith.run(listOf("adopt", "--dry-run", "--json"))
        val report =
            runCatching { AdoptReportMapper.toReport(found.stdout) }.getOrElse {
                ports.notice.warn("Nothing was adopted. ${FailureTextMapper.describe(found)}")
                return false
            }
        if (report.blockers.isNotEmpty()) {
            ports.notice.warn("Nothing was adopted. ${report.blockers.first()}")
            return false
        }
        var move: String? = null
        report.rootProject?.let { project ->
            val choice =
                ports.prompt
                    .pickOne(
                        "$TITLE: where does ${project.name} go?",
                        "${project.manifest} is at the root. smith manages one folder per project and does not pick it for you. Move it into a folder of its own so smith can manage it?",
                        AdoptReportMapper.moveRows(project),
                        "Next",
                    )?.name
                    ?: return false
            move = askMove(ports, project, choice) ?: return false
        }
        val gitattributes =
            pick(
                ports,
                "line endings",
                "Add a .gitattributes with * text=auto eol=lf, so every machine writes the same line endings? A repository that has one keeps it.",
                yesRows("add", "skip", "Add it", "Leave line endings alone"),
            )
                ?: return false
        val agent =
            pick(
                ports,
                "agent files",
                "Add the MonoSmith agent files, so Claude Code can work in this workspace through smith?",
                listOf(
                    ChoiceRowContract("add", "Add them", "(default)", "Writes .claude/agents/smith.md and merges .mcp.json."),
                    ChoiceRowContract("skip", "Not now", "", "You can add them later with smith agent install."),
                    ChoiceRowContract("never", "Never", "", "Records [agent] enabled = false in smith.toml."),
                ),
            ) ?: return false
        val providers = askProviders(ports) ?: return false
        val existing =
            if (report.providers.isEmpty() || providers.isEmpty()) {
                null
            } else {
                pick(
                    ports,
                    "your pipelines",
                    "The repository already has a pipeline for ${report.providers.joinToString(
                        ", ",
                    )}. Replace it with smith's, or keep it?",
                    listOf(
                        ChoiceRowContract("keep", "Keep mine", "(default)", "smith writes no pipeline for that provider."),
                        ChoiceRowContract("replace", "Replace it", "", "smith writes its own; yours is kept as <file>.pre-smith."),
                    ),
                ) ?: return false
            }
        val plan = AdoptReportMapper.initArguments(move, gitattributes, agent, providers, existing, apply = false)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = plan,
                header = "Nothing was changed yet. This is the plan with its diffs. smith writes smith.toml and the pipelines.",
                okText = "Bring it in",
                offerCommit = true,
                apply = { AdoptReportMapper.initArguments(move, gitattributes, agent, providers, existing, apply = true) },
                done = "The folder is a MonoSmith workspace",
                operation = "adopt",
            ),
            commitLater,
        )
    }

    private fun yesRows(
        yes: String,
        no: String,
        yesText: String,
        noText: String,
    ) = listOf(ChoiceRowContract(yes, yesText, "(default)", ""), ChoiceRowContract(no, noText, "", ""))

    private fun pick(
        ports: FlowPortsContract,
        what: String,
        header: String,
        rows: List<ChoiceRowContract>,
    ): String? = ports.prompt.pickOne("$TITLE: $what", "$header The first answer is the default.", rows, "Next")?.name

    private fun askMove(
        ports: FlowPortsContract,
        project: RootProjectContract,
        choice: String,
    ): String? {
        if (choice == "none") return AdoptReportMapper.LEAVE
        val whole = choice == "other"
        val typed =
            ports.prompt.text(
                "$TITLE: ${if (whole) "folder" else "name"}",
                if (whole) "The folder, from the repository root (for example tools/${project.name}):" else "The name in $choice/:",
                if (whole) "" else project.name,
            ) { if (whole) AdoptReportMapper.folderOrProblem(it).second else AdoptReportMapper.nameProblem(it) } ?: return null
        return if (whole) AdoptReportMapper.folderOrProblem(typed).first else "$choice/$typed"
    }

    private fun askProviders(ports: FlowPortsContract): List<String>? {
        val listed = ports.smith.run(listOf("new", "--questions", "--json"))
        val question = runCatching { NewQuestionsMapper.toQuestions(listed.stdout).first { it.key == "ci" } }.getOrNull()
        val rows = question?.let { NewQuestionsMapper.rows(it) } ?: return emptyList()
        return ports.prompt.pickMany("$TITLE: CI", "${question.prompt} Tick none to write no pipeline. The default is ticked.", rows)
    }
}
