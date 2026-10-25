import type { Stack } from '../protocol/index.ts'

/** The stack's plugins other than its language and scaffold, in the order they apply: toolchain, structure, deploy. */
export function stackPlugins (stack: Pick<Stack, 'toolchain' | 'structure' | 'deploy'>): string[] {
  return [...(stack.toolchain ?? []), ...(stack.structure ?? []), ...(stack.deploy ?? [])]
}
