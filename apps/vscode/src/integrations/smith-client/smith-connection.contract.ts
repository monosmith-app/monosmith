/** What `smith serve` answers to `initialize`. */
export interface SmithInitialized {
  name:            string
  version:         string
  protocolVersion: number
  eventVersion:    number
  methods:         string[]
}

/** A JSON-RPC conversation with `smith serve`. */
export interface SmithConnection {
  /**
   * Calls a method and resolves with its result; rejects with the server's
   * message when the call failed, or when the server ended.
   */
  request: <T>(method: string, params?: unknown) => Promise<T>
  /**
   * Listens to the `event` notifications of runs.
   * @returns what stops the listening
   */
  onEvent: (handler: (event: unknown) => void) => () => void
  /** Rejects what is still waiting and ignores later input. */
  close:   (reason?: Error) => void
}

/** How a `smith serve` that nobody stopped ended. */
export interface SmithExit {
  code:   number | null
  signal: NodeJS.Signals | null
  /** The last lines it wrote to stderr. */
  stderr: string
}

/** A smith that is running and has said what it speaks. */
export interface SmithSession {
  connection: SmithConnection
  version:    string
  /** `newer`: smith speaks a later protocol than this extension knows. */
  protocol:   'supported' | 'newer'
  /** Asks the server to end and stops the process. */
  stop:       () => void
}
