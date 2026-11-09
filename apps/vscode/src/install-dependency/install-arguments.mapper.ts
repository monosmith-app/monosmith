import type { InstallRequest } from './install-request.contract.ts'
import { flag } from '../smith-flags/index.ts'

/**
 * The `smith install` arguments for a request. A library is named with `workspace:`, so smith never asks which
 * is meant; a registry package carries `--registry` for the same reason.
 * @param request - what the person chose
 * @returns the arguments, without the program name
 */
export function installArguments (request: InstallRequest): string[] {
  if (request.source === 'workspace') {
    return ['install', flag.project, request.project, `workspace:${request.library ?? ''}`, ...(request.range === undefined ? [] : ['--range', request.range])]
  }

  return ['install', flag.project, request.project, ...(request.dev ? ['-D'] : []), '--registry', ...request.packages]
}

/**
 * Splits what was typed into package names.
 * @param text - names separated by spaces or commas
 * @returns the names
 */
export function parsePackages (text: string): string[] {
  return text.split(/[\s,]+/).filter(Boolean)
}
