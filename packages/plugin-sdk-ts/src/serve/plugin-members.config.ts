import type { SmithPlugin } from './plugin.contract.ts'

type Member = Exclude<keyof SmithPlugin, 'describe'>

/**
 * Which member of a plugin answers each method of the protocol (other than
 * `initialize` and `shutdown`, which the serve loop answers), and which members
 * the plugin must have for the method to count as implemented. A plugin with
 * `install` and no `uninstall` does not implement the `dependencies`
 * capability, as in the Go SDK, where the interface needs both methods. A spec
 * checks the table against the methods of the schema, so a method added to the
 * protocol cannot be left out.
 */
export const PLUGIN_MEMBERS: Readonly<Record<string, { member: Member, requires: readonly Member[] }>> = {
  'discover':               { member: 'discover', requires: ['discover'] },
  'generate':               { member: 'generate', requires: ['generate'] },
  'version/write':          { member: 'writeVersion', requires: ['writeVersion'] },
  'publish':                { member: 'publish', requires: ['publish'] },
  'ci':                     { member: 'ci', requires: ['ci'] },
  'doctor':                 { member: 'doctor', requires: ['doctor'] },
  'contribute':             { member: 'contribute', requires: ['contribute'] },
  'remove':                 { member: 'remove', requires: ['remove'] },
  'project/kind':           { member: 'changeKind', requires: ['changeKind'] },
  'published':              { member: 'published', requires: ['published'] },
  'dependencies/install':   { member: 'install', requires: ['install', 'uninstall'] },
  'dependencies/uninstall': { member: 'uninstall', requires: ['install', 'uninstall'] },
  'dependencies/declared':  { member: 'declared', requires: ['declared', 'rewrite'] },
  'dependencies/rewrite':   { member: 'rewrite', requires: ['declared', 'rewrite'] },
  'dependencies/latest':    { member: 'latest', requires: ['latest'] },
  'sbom/components':        { member: 'components', requires: ['components'] },
}
