import { commandLineOf } from '../command-palette/index.ts'
import type { HomeButton, HomeFacts, HomeItem, HomeRelease } from './home-facts.contract.ts'
import { say, sayCount, type WordKey } from '../../shared/words/index.ts'

/** How many failing checks the detail of the setup item names. */
const MAX_PROBLEMS = 2

/** A button for a command of the extension, with the smith command line the command stands for. */
function button (label: WordKey, command: string, line = commandLineOf(command)): HomeButton {
  return { label: say(label), command, ...(line !== undefined && { line }) }
}

/**
 * What needs the person now, the most urgent first, and only what applies, as the terminal's Home
 * queue (`dashboard.Queue`) lists it: a workspace without projects, the first doctor run that is
 * still going, a doctor that could not run, the last run that failed, the checks that fail and
 * warn, the work the changes left, a release that is ready and the updates. When nothing asks for
 * attention it says so first and offers a build.
 * @param facts - what the extension holds now
 * @returns at least one item; choosing an item runs its first button
 */
export function homeQueue (facts: HomeFacts): HomeItem[] {
  if (facts.projectsLoaded && facts.projects === 0) {
    return [{
      kind:     'empty',
      severity: 'info',
      title:    say('dashboard.empty.title'),
      detail:   say('dashboard.empty.detail'),
      buttons:  [button('dashboard.add.label', 'monosmith.addProject'), button('dashboard.compose.label', 'monosmith.composeStack')],
    }]
  }
  if (facts.healthError === undefined && facts.health === undefined) {
    return [{
      kind:     'checking',
      severity: 'muted',
      title:    say('dashboard.headline.waiting', { projects: sayCount('dashboard.project', facts.projects) }),
      detail:   say('dashboard.checking.detail'),
      buttons:  [],
    }]
  }
  const queue: HomeItem[] = []
  if (facts.healthError !== undefined) {
    queue.push({ kind: 'doctor', severity: 'failure', title: say('dashboard.checkerr.title'), detail: facts.healthError, buttons: [button('dashboard.retry.label', 'monosmith.health.refresh')] })
  }
  const last = facts.last
  if (last !== undefined && !last.running && !last.ok && !last.stopped) {
    queue.push({
      kind:     'run',
      severity: 'failure',
      title:    say('dashboard.run.failed', { what: last.task ?? shortCommand(last.command) }),
      detail:   last.ago,
      buttons:  [button('dashboard.run.open', 'monosmith.runs.open', undefined), button('dashboard.run.again', 'monosmith.rerunLast', last.command)],
    })
  }
  const health = facts.health
  if (health !== undefined && health.failed > 0) {
    const failed = sayCount('dashboard.check.fails', health.failed)
    queue.push({
      kind:     'setup',
      severity: 'failure',
      title:    health.fixable > 0 ? say('dashboard.setup.title', { failed, fixable: sayCount('dashboard.setup.fixable', health.fixable) }) : failed,
      detail:   health.problems.slice(0, MAX_PROBLEMS).join(' · '),
      buttons:  [...(health.fixable > 0 ? [button('dashboard.fix.label', 'monosmith.health.fix')] : []), button('dashboard.doctor.label', 'monosmith.health.focus')],
    })
  }
  if (health !== undefined && health.warnings > 0) {
    queue.push({
      kind:     'warnings',
      severity: 'warning',
      title:    sayCount('dashboard.warn.title', health.warnings, { first: health.warningNotes[0] ?? '' }),
      buttons:  [...(health.fixable > 0 ? [button('dashboard.fix.label', 'monosmith.health.fix')] : []), button('dashboard.doctor.label', 'monosmith.health.focus')],
    })
  }
  if (facts.affected > 0) {
    queue.push({ kind: 'changes', severity: 'warning', title: sayCount('dashboard.changes.title', facts.affected), buttons: [button('dashboard.changes.test', 'monosmith.testChanged')] })
  }
  if (facts.release !== undefined) {
    queue.push({ kind: 'release', severity: 'info', title: releaseTitle(facts.release), buttons: [button('dashboard.release.open', 'monosmith.release.focus')] })
  }
  if (facts.outdated === undefined) {
    queue.push({ kind: 'updates', severity: 'muted', title: say('dashboard.updates.unlooked'), buttons: [button('dashboard.updates.review', 'monosmith.dependencies.focus')] })
  } else if (facts.outdated > 0) {
    queue.push({ kind: 'updates', severity: 'info', title: sayCount('dashboard.outdated', facts.outdated), buttons: [button('dashboard.updates.review', 'monosmith.dependencies.focus')] })
  }
  if (queue.every((item) => item.severity === 'muted')) {
    const good: HomeItem = { kind: 'good', severity: 'success', title: say('dashboard.good.title'), buttons: [button('dashboard.good.build', 'monosmith.buildAll')] }
    if (last !== undefined && last.ok && last.took !== '') good.detail = say('dashboard.good.last', { command: shortCommand(last.command), took: last.took })
    queue.unshift(good)
  }

  return queue
}

/** Says how many packages leave and names one. */
function releaseTitle (release: HomeRelease): string {
  const bump = release.from === ''
    ? say('dashboard.release.first', { name: release.name, to: release.to })
    : say('dashboard.release.bump', { name: release.name, from: release.from, to: release.to })
  const packages = sayCount('dashboard.release.packages', release.packages)

  return release.commits === 0
    ? say('dashboard.release.title', { packages, bump })
    : say('dashboard.release.title_commits', { packages, bump, commits: sayCount('dashboard.release.commits', release.commits) })
}

/** A command line without `smith run `, which every task run has. */
function shortCommand (command: string): string {
  return command.replace(/^smith /, '').replace(/^run /, '')
}
