import type { Peer } from '../protocol/index.ts'

/** What the host tells the plugin, and how long it waits. */
export interface HostOptions {
  /** The absolute workspace root sent in `initialize`. A plugin that reads manifests resolves them against it. */
  workspaceRoot:        string
  /** The host's name and version. Default: `smith`, `dev`. */
  host?:                Peer
  /** How long `initialize` may take. Default: 30 seconds. */
  initializeTimeoutMs?: number
}

/** Options of {@link startPlugin}: a plugin process to launch. */
export interface StartOptions extends HostOptions {
  /** The plugin's argv: the program and its arguments. */
  command: string[]
  /** The environment. Default: the current one. `SMITH_PLUGIN=1` is set in it unless it already has a value, as smith does. */
  env?:    NodeJS.ProcessEnv
  /** The plugin's working directory. Default: the current one. */
  cwd?:    string
  /** What happens to the plugin's standard error, where it logs. Default: it goes to this process's standard error. */
  stderr?: 'inherit' | 'ignore'
}

/** The longest a call may take unless the caller says otherwise: generate and publish run real toolchains. */
export const DEFAULT_CALL_TIMEOUT_MS = 2 * 60 * 1000

/** Options of one {@link PluginClient.call}. */
export interface CallOptions {
  /** How long the plugin may take. On expiry the plugin process is killed: its stream can no longer be trusted. Default: 2 minutes. */
  timeoutMs?: number
}
