import { cpSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { SCHEMA_PATH, generateProtocol } from './generate-protocol.use-case.ts'

const packageRoot = fileURLToPath(new URL('../../', import.meta.url))

// The drift guard: the types the SDK ships are the ones the schema produces, and the schema is the Go SDK's.
describe('generateProtocol', () => {
  it('finds the committed types and schema module current', () => {
    expect(generateProtocol({ packageRoot, check: true }), 'run `smith run generate -p plugin-sdk-ts` and commit the result').toEqual([])
  })

  it('reports what is stale with --check and rewrites it without', () => {
    const copy = mkdtempSync(join(tmpdir(), 'sdk-generate-'))
    try {
      cpSync(join(packageRoot, 'src', 'protocol'), join(copy, 'src', 'protocol'), { recursive: true })
      cpSync(join(packageRoot, SCHEMA_PATH), join(copy, SCHEMA_PATH), { recursive: true })
      const types = join(copy, 'src', 'protocol', 'protocol-types.contract.ts')
      const original = readFileSync(types, 'utf8')
      writeFileSync(types, `${original}// edited by hand\n`)
      expect(generateProtocol({ packageRoot: copy, check: true })).toEqual(['src/protocol/protocol-types.contract.ts'])
      expect(readFileSync(types, 'utf8')).toContain('edited by hand')
      expect(generateProtocol({ packageRoot: copy })).toEqual(['src/protocol/protocol-types.contract.ts'])
      expect(readFileSync(types, 'utf8')).toBe(original)
      expect(generateProtocol({ packageRoot: copy, check: true })).toEqual([])
    } finally {
      rmSync(copy, { recursive: true, force: true })
    }
  })
})
