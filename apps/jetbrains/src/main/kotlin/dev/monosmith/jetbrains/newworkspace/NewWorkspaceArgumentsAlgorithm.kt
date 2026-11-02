package dev.monosmith.jetbrains.newworkspace

import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/**
 * What the person decides about a new workspace: the questions of `smith new`, one field each.
 *
 * @property ci where the tests run on every push: `github`, `azure`, `gitlab`, `circleci` or `bitbucket`
 * @property language the main language, or `later`
 * @property nodeStack the tests and lint of Node projects (`node-jest`, `node-vitest`, `node-ts`); used only for `node`
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

/** The command line and the lists of `smith new`, as the CLI documents them (docs/reference/cli.md). */
object NewWorkspaceArgumentsAlgorithm {
    /** The CI systems `--ci` accepts. */
    val ciRows =
        listOf(
            ChoiceRowContract("github", "GitHub Actions", "(default)", "Pipelines in .github/workflows.", picked = true),
            ChoiceRowContract("azure", "Azure Pipelines", "", "azure-pipelines.yml"),
            ChoiceRowContract("gitlab", "GitLab CI", "", ".gitlab-ci.yml"),
            ChoiceRowContract("circleci", "CircleCI", "", ".circleci/config.yml"),
            ChoiceRowContract("bitbucket", "Bitbucket Pipelines", "", "bitbucket-pipelines.yml"),
        )

    /** The main languages `--language` accepts. */
    val languageRows =
        listOf(
            ChoiceRowContract("later", "Decide later", "(default)", "No language yet; you pick one for each project you add."),
            ChoiceRowContract("node", "TypeScript / JavaScript", "", "node"),
            ChoiceRowContract("python", "Python", "", "python"),
            ChoiceRowContract("go", "Go", "", "go"),
            ChoiceRowContract("dotnet", "C# / .NET", "", "dotnet"),
            ChoiceRowContract("jvm", "Java / Kotlin", "", "jvm"),
            ChoiceRowContract("rust", "Rust", "", "rust"),
        )

    /** The tests and lint stacks of Node projects. */
    val nodeStackRows =
        listOf(
            ChoiceRowContract("node-jest", "Jest", "(default)", "node-jest"),
            ChoiceRowContract("node-vitest", "Vitest", "", "node-vitest"),
            ChoiceRowContract("node-ts", "TypeScript only", "", "node-ts"),
        )

    /** The code layouts `--architecture` accepts. */
    val architectureRows =
        listOf(
            ChoiceRowContract("vertical-slices", "Vertical slices", "(default)", "One folder per outcome, files named by role."),
            ChoiceRowContract("hexagonal", "Hexagonal", "", "Ports and adapters."),
            ChoiceRowContract("clean", "Clean", "", "Clean architecture layers."),
            ChoiceRowContract("ddd", "Domain-driven design", "", "Domain, application, infrastructure."),
            ChoiceRowContract("modular-monolith", "Modular monolith", "", "One module per business area."),
            ChoiceRowContract("mvc", "MVC", "", "Models, views, controllers."),
            ChoiceRowContract("package-by-layer", "Package by layer", "", "One folder per technical layer."),
            ChoiceRowContract("none", "No layout check", "", "smith checks nothing about the folders."),
        )

    /** The rows of a yes or no question; [yes] says what yes does, [no] what no does; yes is the default when [defaultYes]. */
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
            addAll(listOf("new", choices.name, "--here", "--yes", "--ci", choices.ci.joinToString(","), "--language", choices.language))
            if (choices.language == "node") addAll(listOf("--default-stack", "node=${choices.nodeStack}"))
            addAll(listOf("--architecture", choices.architecture))
            add(if (choices.agent) "--agent" else "--no-agent")
            add(if (choices.devcontainer) "--devcontainer" else "--no-devcontainer")
            if (!choices.commit) add("--no-commit")
            if (preview) add("--dry-run")
        }

    private fun label(
        rows: List<ChoiceRowContract>,
        value: String,
    ): String = rows.firstOrNull { it.name == value }?.label ?: value

    /** What the choices mean, one line each, for the summary shown before smith writes anything. */
    fun describe(
        choices: NewWorkspaceChoicesContract,
        folderName: String,
    ): List<String> =
        listOf(
            "Name: ${choices.name}${if (choices.name == folderName) " (the name of this folder)" else ""}",
            "CI: ${choices.ci.joinToString(", ") { label(ciRows, it) }}",
            "Main language: ${label(languageRows, choices.language)}",
        ) +
            (
                if (choices.language ==
                    "node"
                ) {
                    listOf("Tests and lint for Node projects: ${label(nodeStackRows, choices.nodeStack)}")
                } else {
                    emptyList()
                }
            ) +
            listOf(
                "Code layout smith checks: ${label(architectureRows, choices.architecture)}",
                "Agent files for Claude Code: ${if (choices.agent) "yes" else "no"}",
                "Dev container: ${if (choices.devcontainer) "yes" else "no"}",
                "First commit of smith's own files: ${if (choices.commit) "yes, smith makes it" else "no, the files stay uncommitted"}",
            )
}
