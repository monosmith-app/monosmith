package dev.monosmith.jetbrains.smithcommand

/**
 * Generated from libs/actions/flagvocabulary/flag_vocabulary.json by `go run ./flaggen` in libs/actions. Do not edit: change the vocabulary and run it again.
 */
object SmithFlagsConfig {
    /** The project or projects a command acts on; repeat it or separate names with commas (in smith the workspace is the whole repository, so -w is only a synonym). */
    const val PROJECT = "-p"

    /** The long form of [PROJECT]. */
    const val PROJECT_LONG = "--project"

    /** A programming language, by the name its plugin has. */
    const val LANG = "-l"

    /** The long form of [LANG]. */
    const val LANG_LONG = "--lang"

    /** The file or folder a command writes its result to. */
    const val OUTPUT = "-o"

    /** The long form of [OUTPUT]. */
    const val OUTPUT_LONG = "--output"

    /** Do not ask for confirmation: take the safe answer. */
    const val YES = "-y"

    /** The long form of [YES]. */
    const val YES_LONG = "--yes"

    /** Change the manifests but do not run the package manager. */
    const val NO_INSTALL = "--no-install"

    /** The long form of [NO_INSTALL]. */
    const val NO_INSTALL_LONG = "--no-install"

    /** Show what the command would do and change nothing. */
    const val DRY_RUN = "--dry-run"

    /** The long form of [DRY_RUN]. */
    const val DRY_RUN_LONG = "--dry-run"

    /** Print the result as JSON. */
    const val JSON = "--json"

    /** The long form of [JSON]. */
    const val JSON_LONG = "--json"

    /** Commit the change as one operation so smith undo can revert it. */
    const val COMMIT = "--commit"

    /** The long form of [COMMIT]. */
    const val COMMIT_LONG = "--commit"

    /** Build the optimised profile (cargo --release, dotnet -c Release); a tool without profiles ignores it. */
    const val RELEASE = "--release"

    /** The long form of [RELEASE]. */
    const val RELEASE_LONG = "--release"
}
