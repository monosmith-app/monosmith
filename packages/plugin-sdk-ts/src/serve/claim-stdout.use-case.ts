import type { FrameSink } from './protocol-streams.contract.ts'

/**
 * Takes the process's standard output for the protocol and nothing else. A
 * `console.log`, a library's `process.stdout.write` or a child process's
 * inherited output would corrupt the stream the host reads, so from here on
 * everything written to `process.stdout` goes to standard error, where the
 * host shows it to the user. The returned sink is the only way to the real
 * standard output.
 */
export function claimStdout (): FrameSink {
  const stdout = process.stdout
  const write = stdout.write.bind(stdout) as FrameSink['write']
  stdout.write = ((...args: Parameters<typeof process.stderr.write>) => process.stderr.write(...args)) as typeof stdout.write

  return { write }
}
