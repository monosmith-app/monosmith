import type { RunEvent } from './run-event.contract.ts'

/**
 * The line an event adds to the output panel, or undefined for an event that
 * adds none. The command line comes first, so what ran can be seen and
 * repeated in a terminal.
 */
export function describeEvent (event: RunEvent): string | undefined {
  switch (event.event) {
    case 'run.started': {
      return `$ ${event.command ?? 'smith run'}`
    }
    case 'task.output': {
      return `${event.task ?? ''} | ${event.line ?? ''}`
    }
    case 'task.finished': {
      const took = (event.durationMs ?? 0) > 0 ? ` in ${((event.durationMs ?? 0) / 1000).toFixed(1)}s` : ''
      const why = event.reason === undefined ? '' : ` (${event.reason})`

      return `${event.task ?? ''} ${event.status ?? 'failed'}${took}${why}`
    }
    case 'run.finished': {
      const s = event.summary
      const counts = s === undefined ? '' : `: ${s.passed} passed, ${s.cached} cached, ${s.failed} failed, ${s.skipped} skipped`

      return `${event.ok === true ? 'Run finished' : 'Run failed'}${counts}`
    }
    default: {
      return undefined
    }
  }
}
