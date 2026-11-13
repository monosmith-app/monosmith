// Generated from libs/actions/flagvocabulary/flag_vocabulary.json by `go run ./flaggen` in libs/actions. Do not edit: change the vocabulary and run it again.

/** The spelling of each shared flag a smith command line should use: the shorthand when there is one. */
export const flag = {
  project:   '-p',
  lang:      '-l',
  output:    '-o',
  yes:       '-y',
  noInstall: '--no-install',
  dryRun:    '--dry-run',
  json:      '--json',
  commit:    '--commit',
  release:   '--release',
  fix:       '--fix',
} as const

/** The long form of each shared flag. */
export const longFlag = {
  project:   '--project',
  lang:      '--lang',
  output:    '--output',
  yes:       '--yes',
  noInstall: '--no-install',
  dryRun:    '--dry-run',
  json:      '--json',
  commit:    '--commit',
  release:   '--release',
  fix:       '--fix',
} as const
