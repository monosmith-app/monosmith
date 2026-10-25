import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import type { SchemaDocument } from '../../src/protocol/schema-document.contract.ts'
import { mapSchemaToModule } from './schema-module.mapper.ts'
import { mapSchemaToTypes } from './typescript-types.mapper.ts'

export interface GenerateOptions {
  /** The package's root directory (packages/plugin-sdk-ts). */
  packageRoot: string
  /** Write nothing; report the files that differ from what would be written. */
  check?:      boolean
}

export interface GeneratedFile {
  /** Relative to the package root, forward slashes. */
  path:    string
  content: string
}

/** The schema, relative to the package root: the Go SDK's, next to the types it comes from. */
export const SCHEMA_PATH = '../plugin-sdk/schema/protocol.schema.json'

/** What the generator writes, from the schema's text. */
export function generatedFiles (schemaText: string): GeneratedFile[] {
  const schema = JSON.parse(schemaText) as SchemaDocument

  return [
    { path: 'src/protocol/protocol-types.contract.ts', content: mapSchemaToTypes(schema) },
    { path: 'src/protocol/protocol-schema.contract.ts', content: mapSchemaToModule(schema) },
  ]
}

/**
 * Writes the generated files, or, with `check`, returns the ones that are not
 * what the schema produces. An empty result from a check means they are current.
 */
export function generateProtocol (options: GenerateOptions): string[] {
  const schemaText = readFileSync(join(options.packageRoot, SCHEMA_PATH), 'utf8')
  const stale: string[] = []
  for (const file of generatedFiles(schemaText)) {
    const target = join(options.packageRoot, file.path)
    let current: string | undefined
    try {
      current = readFileSync(target, 'utf8').replaceAll('\r\n', '\n')
    } catch {
      current = undefined
    }
    if (current === file.content) continue
    stale.push(file.path)
    if (options.check === true) {
      continue
    }

    mkdirSync(dirname(target), { recursive: true })
    writeFileSync(target, file.content)
  }

  return stale
}
