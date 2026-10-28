import { optOutArguments } from './install-arguments.mapper.ts'
import { sayProblem } from './install-outcome.mapper.ts'
import type { SmithRun } from './install-run.contract.ts'

/**
 * Records `[agent] enabled = false` in smith.toml through `smith config set`,
 * so smith doctor stops asking for the agent files.
 * @param run - runs smith in the workspace folder
 * @returns undefined when it is recorded, otherwise what smith said
 */
export async function optOutOfAgent (run: (args: string[]) => Promise<SmithRun>): Promise<string | undefined> {
  const result = await run(optOutArguments())

  return result.code === 0 ? undefined : (sayProblem(result.stderr) || `smith stopped with exit code ${result.code}`)
}
