import type { MissingTool } from './add-options.contract.ts'

/** What a person calls the programs smith checks for. */
const TITLES: Record<string, string> = {
  dotnet:  'the .NET SDK',
  go:      'Go',
  node:    'Node.js',
  npm:     'Node.js',
  npx:     'Node.js',
  python:  'Python',
  python3: 'Python',
  pip:     'Python',
  cargo:   'Rust',
  rustc:   'Rust',
  java:    'Java (JDK)',
  javac:   'Java (JDK)',
  gradle:  'Gradle',
  mvn:     'Maven',
  flutter: 'Flutter',
  dart:    'Dart',
  git:     'Git',
  docker:  'Docker',
  cmake:   'CMake',
}

/**
 * Reads the output of a failed `smith add` and tells whether it failed because
 * a tool is missing, which leaves the new project on disk.
 * @param output - what smith printed, standard output then standard error
 * @returns the tool and what was made, or undefined for any other failure
 */
export function parseMissingTool (output: string): MissingTool | undefined {
  const program = /(\S+) is not installed or not on PATH/.exec(output)?.[1]?.toLowerCase().replace(/\.exe$/, '')
  if (program === undefined) return undefined

  const url = /Install it from (https?:\/\/[^\s,]+)/.exec(output)?.[1]
  const dir = /^Creating \S+ \S+ in (\S+?):/m.exec(output)?.[1]?.replace(/\/$/, '')
  const files = Number(/Wrote (\d+) file/.exec(output)?.[1] ?? 0)
  const check = /^Verifying: smith (run .+ -p \S+)$/m.exec(output)?.[1]?.split(' ')

  return { title: TITLES[program] ?? program, files, ...(url !== undefined && { url }), ...(dir !== undefined && { dir }), ...(check !== undefined && { check }) }
}
