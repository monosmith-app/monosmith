import { readFileSync, writeFileSync } from 'node:fs'
import { JsonDocument } from '@monosmith/plugin-sdk'

/** The file that makes a directory a widget project. */
export const MANIFEST_FILE = 'widget.json'

/** The content of a widget.json. */
export interface Manifest {
  name:        string
  version:     string
  /** app, lib or package; default lib */
  kind:        string
  description: string
  license:     string
  /** names of other widgets this one imports */
  uses:        string[]
  stack:       string
}

const SEMVER = /^\d+\.\d+\.\d+$/

function slashed (path: string): string {
  return path.replaceAll('\\', '/')
}

/** Reads a widget.json. A member it does not have is empty. */
export function readManifest (path: string): Manifest {
  let parsed: Partial<Manifest>
  try {
    parsed = JSON.parse(readFileSync(path, 'utf8')) as Partial<Manifest>
  } catch (error) {
    throw new Error(`${slashed(path)}: ${error instanceof Error ? error.message : String(error)}`, { cause: error })
  }

  return {
    name:        parsed.name ?? '',
    version:     parsed.version ?? '',
    kind:        parsed.kind ?? '',
    description: parsed.description ?? '',
    license:     parsed.license ?? '',
    uses:        parsed.uses ?? [],
    stack:       parsed.stack ?? '',
  }
}

/** What the test task checks: a name and an x.y.z version. */
export function validate (manifest: Manifest): void {
  if (manifest.name === '') throw new Error(`${MANIFEST_FILE} has no name`)
  if (!SEMVER.test(manifest.version)) throw new Error(`${MANIFEST_FILE}: version "${manifest.version}" is not x.y.z`)
}

/** The widget.json with one string member set, every other key keeping its order and layout. */
export function withField (path: string, key: string, value: string): string {
  const document = JsonDocument.parse(readFileSync(path, 'utf8'))
  document.setString(key, value)

  return document.render()
}

/** Rewrites one string member of a widget.json. */
export function setField (path: string, key: string, value: string): void {
  writeFileSync(path, withField(path, key, value))
}
