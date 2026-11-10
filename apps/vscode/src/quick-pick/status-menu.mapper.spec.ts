import { describe, expect, it } from 'vitest'
import { statusMenuEntries } from './status-menu.mapper.ts'

const idle = { running: false, hasLastRun: false, failed: 0 }

describe('statusMenuEntries', () => {
  it('offers run affected, doctor, refresh, the agent, output and settings', () => {
    const commands = statusMenuEntries(idle).map((entry) => entry.command)
    expect(commands).toEqual(expect.arrayContaining(['monosmith.runAffected', 'monosmith.doctor', 'monosmith.refresh', 'monosmith.showOutput', 'monosmith.openSettings', 'monosmith.talkToAgent']))
  })

  it('offers to add a project, open the report and undo the last change', () => {
    const commands = statusMenuEntries(idle).map((entry) => entry.command)
    expect(commands).toEqual(expect.arrayContaining(['monosmith.addProject', 'monosmith.openReport', 'monosmith.undo']))
  })

  it('offers cancel only while a run goes, and first', () => {
    expect(statusMenuEntries(idle).some((entry) => entry.command === 'monosmith.cancelRun')).toBe(false)
    expect(statusMenuEntries({ ...idle, running: true })[0]?.command).toBe('monosmith.cancelRun')
  })

  it('offers re-run last only once something has run', () => {
    expect(statusMenuEntries(idle).some((entry) => entry.command === 'monosmith.rerunLast')).toBe(false)
    expect(statusMenuEntries({ ...idle, hasLastRun: true }).some((entry) => entry.command === 'monosmith.rerunLast')).toBe(true)
  })

  it('says how many tasks failed on the doctor entry', () => {
    const doctor = statusMenuEntries({ ...idle, failed: 2 }).find((entry) => entry.command === 'monosmith.doctor')
    expect(doctor?.description).toContain('2 tasks failed')
  })
})

describe('statusMenuEntries for a folder that can be converted', () => {
  it('offers the conversion and its preview instead of runs', () => {
    const commands = statusMenuEntries({ ...idle, view: 'adoptable' }).map((entry) => entry.command)
    expect(commands).toEqual(['monosmith.adopt.start', 'monosmith.adopt.preview', 'monosmith.showOutput', 'monosmith.openSettings'])
  })
})

describe('statusMenuEntries for a folder with no smith.toml and nothing to convert', () => {
  it('offers a new workspace in an empty folder, and no runs', () => {
    const commands = statusMenuEntries({ ...idle, view: 'empty' }).map((entry) => entry.command)
    expect(commands).toEqual(['monosmith.new', 'monosmith.showOutput', 'monosmith.openSettings'])
  })

  it('offers to bring in a folder with files, and its preview', () => {
    const entries = statusMenuEntries({ ...idle, view: 'unrecognised' })
    expect(entries.map((entry) => entry.command)).toEqual(['monosmith.adopt.start', 'monosmith.adopt.preview', 'monosmith.showOutput', 'monosmith.openSettings'])
    expect(entries[0]?.label).toContain('Bring this folder into MonoSmith')
  })
})
