package dev.monosmith.jetbrains.smithrun

import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHRUN_SMITH_STOPPED_WITH_EXIT_CODE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Words why a smith command failed, in smith's own words and with its fix. */
object FailureTextMapper {
    /**
     * Describes a run that ended with a code other than 0. smith writes a problem on standard error; a run
     * with none is described by standard output, and a run with neither by its exit code.
     *
     * @param run the failed run
     * @return the problem, then the first fix as `Try: ...`
     */
    fun describe(run: SmithRunContract): String =
        ProblemTextMapper
            .toLine(
                run.stderr.ifBlank { run.stdout },
            ).ifBlank { say(JETBRAINS_SMITHRUN_SMITH_STOPPED_WITH_EXIT_CODE, run.code.toString()) }
}
