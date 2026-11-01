import type { SwapRequest } from './swap-choice.contract.ts'

/**
 * The arguments of `smith swap` for a request.
 * @param request - what the person chose
 * @param mode - `plan`: `--dry-run`, smith changes nothing; `apply`: `--yes`, smith writes without a question (the person already agreed)
 * @param commit - with `apply`: `--commit`, so smith makes the swap one commit that `smith undo` can revert; without it smith commits nothing
 */
export function swapArguments (request: SwapRequest, mode: 'plan' | 'apply', commit = false): string[] {
  return [
    'swap', request.project, request.axis, request.choice,
    mode === 'plan' ? '--dry-run' : '--yes',
    ...(mode === 'apply' && commit ? ['--commit'] : []),
  ]
}
