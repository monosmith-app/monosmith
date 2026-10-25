/** Where the serve loop writes its frames: a Node `Writable` fits. */
export interface FrameSink {
  write: (chunk: Buffer, callback: (error?: Error | null) => void) => unknown
}

/** The streams {@link serveStreams} runs on. */
export interface ServeStreams {
  /** The host's requests, as bytes (a Node `Readable` fits). */
  input:  AsyncIterable<Uint8Array>
  /** The plugin's responses. */
  output: FrameSink
  /** Where a warning about the plugin goes. Default: standard error. */
  log?:   (message: string) => void
}
