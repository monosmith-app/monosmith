/** Remembers which warnings were shown, so each is shown once per session. */
export interface WarnedOnceStore {
  /** True the first time a key is asked for, false after. */
  firstTime: (key: string) => boolean
}

/** A store that has warned about nothing yet. */
export function createWarnedOnceStore (): WarnedOnceStore {
  const seen = new Set<string>()

  return {
    firstTime: (key) => {
      if (seen.has(key)) return false
      seen.add(key)

      return true
    },
  }
}
