/** What the status bar item shows and says. */
export interface StatusText {
  text:    string
  tooltip: string
}

/**
 * The status bar text: the smith that is running, and how many projects the
 * changes affect when that is known.
 * @param version - smith's version, or undefined while it is not running
 * @param affected - the number of affected projects, if known
 * @param running - whether a run is in progress
 */
export function describeStatus (version: string | undefined, affected: number | undefined, running: boolean): StatusText {
  if (version === undefined) {
    return { text: '$(circle-slash) smith', tooltip: 'MonoSmith: smith was not found or has stopped. Click to try again.' }
  }
  const parts = [`smith ${version}`]
  if (affected !== undefined) parts.push(`${affected} affected`)

  return {
    text:    `${running ? '$(sync~spin)' : '$(tools)'} ${parts.join(' · ')}`,
    tooltip: running ? 'MonoSmith: a run is in progress. Click to refresh.' : 'MonoSmith: click to refresh the workspace.',
  }
}
