package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.pipelinetemplate.AddPipelineUseCase

/** Adds a ready-made pipeline, after the plan. */
class AddPipelineActionHandler : SmithActionHandler({ ide, root -> AddPipelineUseCase.start(ide, root) })
