import type { SmithLocation, SmithSource } from '../smith-client/index.ts'
import { say } from '../words/index.ts'

/** Names where a smith came from, as a phrase that follows "from". */
export function sourceName (source: SmithSource): string {
  switch (source) {
    case 'setting': {
      return say('vscode.smith-version.the-monosmith-path-setting')
    }
    case 'workspace': {
      return say('vscode.smith-version.this-workspace-node-modules-bin-or')
    }
    case 'path': {
      return say('vscode.smith-version.the-path')
    }
  }
}

/** What is known about the smith in use. */
export interface OriginFacts {
  location:         SmithLocation
  version:          string
  /** The workspace had its own smith, left alone because the folder is not trusted. */
  skippedUntrusted: boolean
  /** The version a terminal would run, when it differs and is known. */
  terminalVersion?: string
}

/**
 * The lines that tell which smith the editor runs and why: the version, the
 * path, where it came from, and anything left out on purpose.
 */
export function describeOrigin (facts: OriginFacts): string[] {
  const lines = [
    `smith ${facts.version} from ${sourceName(facts.location.source)}`,
    facts.location.executable,
  ]
  if (facts.skippedUntrusted) {
    lines.push(say('vscode.smith-version.this-workspace-has-its-own-smith'))
  }
  if (facts.terminalVersion !== undefined && facts.terminalVersion !== facts.version) {
    lines.push(say('vscode.smith-version.a-terminal-runs-smith-from-the', { terminalVersion: facts.terminalVersion }))
  }

  return lines
}
