// Writing the generated C# files from the schema, or checking that they are current.

import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { mapSchemaToClientCalls } from './csharp-client.mapper.mjs'
import { mapSchemaToPluginContract, mapSchemaToPluginMethods } from './csharp-plugin.mapper.mjs'
import { mapSchemaToTypes } from './csharp-types.mapper.mjs'

/** The schema, relative to the package root: the Go SDK's, next to the types it comes from. */
export const SCHEMA_PATH = '../plugin-sdk/schema/protocol.schema.json'

/** What the generator writes, from the schema's text: `{ path, content }`, the path relative to the package root. */
export function generatedFiles(schemaText) {
  const schema = JSON.parse(schemaText)
  return [
    { path: 'Protocol/ProtocolTypesContract.cs', content: mapSchemaToTypes(schema) },
    { path: 'Serve/PluginContract.cs', content: mapSchemaToPluginContract(schema) },
    { path: 'Serve/PluginMethodsContract.cs', content: mapSchemaToPluginMethods(schema) },
    { path: 'Host/PluginClientCallsContract.cs', content: mapSchemaToClientCalls(schema) },
  ]
}

/**
 * Writes the generated files, or, with `check`, only lists the ones that are not what the schema produces.
 * An empty result from a check means they are current.
 */
export function generateProtocol({ packageRoot, check = false }) {
  const schemaText = readFileSync(join(packageRoot, SCHEMA_PATH), 'utf8')
  const stale = []
  for (const file of generatedFiles(schemaText)) {
    const target = join(packageRoot, file.path)
    let current
    try {
      current = readFileSync(target, 'utf8').replaceAll('\r\n', '\n')
    } catch {
      current = undefined
    }
    if (current === file.content) continue
    stale.push(file.path)
    if (check) continue
    mkdirSync(dirname(target), { recursive: true })
    writeFileSync(target, file.content)
  }
  return stale
}
