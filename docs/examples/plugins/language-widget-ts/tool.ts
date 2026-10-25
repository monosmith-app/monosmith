import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { MANIFEST_FILE, readManifest, validate } from './manifest.ts'

/**
 * The command line the plugin's tasks call: the tasks run this same program,
 * so the example needs no other tool installed.
 *
 *     plugin.ts check            validate ./widget.json
 *     plugin.ts build <out-dir>  write <out-dir>/widget.txt
 */
export function runTool (args: string[]): void {
  const [command, directory] = args
  if (command === undefined) throw new Error('usage: plugin.ts check | build <out-dir> (smith runs this program as a plugin)')
  const manifest = readManifest(MANIFEST_FILE)
  if (command === 'check') {
    validate(manifest)
    console.log(`${manifest.name} ${manifest.version} is valid`)
  } else if (command === 'build') {
    if (directory === undefined || args.length !== 2) throw new Error('build takes the output directory')
    mkdirSync(directory, { recursive: true })
    const out = join(directory, 'widget.txt')
    writeFileSync(out, `${manifest.name} ${manifest.version}\n`)
    console.log('wrote', out.replaceAll('\\', '/'))
  } else {
    throw new Error(`unknown command "${command}"`)
  }
}
