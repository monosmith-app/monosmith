package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import java.nio.file.Path

/** Reads the plan of a release for the Release view (`smith release --dry-run --json`). It changes nothing. */
object ReadReleasePreviewUseCase {
    /** Blocks: call it off the UI thread. */
    fun read(root: Path): Result<ReleasePreviewContract> =
        runCatching {
            val run = SmithRunClient.run(null, root, listOf("release", "--dry-run", "--json"))
            runCatching { ReleasePreviewMapper.toPreview(run.stdout) }.getOrElse { error(FailureTextMapper.describe(run)) }
        }
}
