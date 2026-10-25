import { buildDirFromProject } from '../build-dir/index.ts'
import type { Task } from '../protocol/index.ts'

/**
 * A task the host can run, hash and cache. The command is an argv array and is
 * never passed through a shell. Other members are the options.
 *
 * ```ts
 * task(['node', 'check.js'], { inputs: ['widget.json'], cache: true, description: 'check the manifest' })
 * ```
 */
export function task (command: string[], options: Omit<Task, 'command'> = {}): Task {
  return { command, ...options }
}

/**
 * The `build` task of a project: cached, built into the project's directory
 * under `dist/` (which is also its `outputs`, restored on a cache hit), and run
 * after the build of every project this one imports (`^build`). `outputDir` is
 * the directory the command is told to write into, `../../dist/libs/core` for
 * `libs/core`, to pass in its argv.
 */
export function buildTask (root: string, command: (outputDir: string) => string[], options: Omit<Task, 'command'> = {}): Task {
  const outputDir = buildDirFromProject(root)

  return { command: command(outputDir), outputs: [outputDir], dependsOn: ['^build'], cache: true, ...options }
}
