import type { SmithLocation, SmithSource } from '../smith-client/index.ts'

/** Names where a smith came from, as a phrase that follows "from". */
export function sourceName (source: SmithSource): string {
  switch (source) {
    case 'setting': {
      return 'the monosmith.path setting'
    }
    case 'workspace': {
      return 'this workspace (node_modules/.bin or .smith/bin)'
    }
    case 'path': {
      return 'the PATH'
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
    lines.push('This workspace has its own smith, which is not used because the workspace is not trusted.')
  }
  if (facts.terminalVersion !== undefined && facts.terminalVersion !== facts.version) {
    lines.push(`A terminal runs smith ${facts.terminalVersion} from the PATH.`)
  }

  return lines
}
