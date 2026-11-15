package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.pipelinetemplate.AddPipelineUseCase

/** Adds a ready-made pipeline, after the plan. */
class AddPipelineActionHandler : SmithActionHandler({ ide, root -> AddPipelineUseCase.start(ide, root) })
