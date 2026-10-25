// The command line of the generator:
//
//   node tools/generate-protocol/generate-protocol.handler.ts            write the files
//   node tools/generate-protocol/generate-protocol.handler.ts --check    fail when they are stale
//
// Run it with `smith run generate -p plugin-sdk-ts` after the Go SDK's schema
// changes (`smith run schema -p plugin-sdk`).
import { fileURLToPath } from 'node:url'
import { generateProtocol } from './generate-protocol.use-case.ts'

const packageRoot = fileURLToPath(new URL('../../', import.meta.url))
const check = process.argv.includes('--check')
const changed = generateProtocol({ packageRoot, check })

if (check && changed.length > 0) {
  console.error(`stale: ${changed.join(', ')}\nrun \`smith run generate -p plugin-sdk-ts\` and commit the result`)
  process.exitCode = 1
} else {
  console.log(changed.length === 0 ? 'the generated files are current' : `wrote ${changed.join(', ')}`)
}
