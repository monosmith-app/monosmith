import { describe, expect, it } from 'vitest'
import { statusMenuEntries } from './status-menu.mapper.ts'

const idle = { running: false, hasLastRun: false, failed: 0 }

describe('statusMenuEntries', () => {
  it('offers run affected, doctor, refresh, output and settings', () => {
    const commands = statusMenuEntries(idle).map((entry) => entry.command)
    expect(commands).toEqual(expect.arrayContaining(['monosmith.runAffected', 'monosmith.doctor', 'monosmith.refresh', 'monosmith.showOutput', 'monosmith.openSettings']))
  })

  it('offers cancel only while a run goes, and first', () => {
    expect(statusMenuEntries(idle).some((entry) => entry.command === 'monosmith.cancelRun')).toBe(false)
    expect(statusMenuEntries({ ...idle, running: true })[0]?.command).toBe('monosmith.cancelRun')
  })

  it('says how many tasks failed on the doctor entry', () => {
    const doctor = statusMenuEntries({ ...idle, failed: 2 }).find((entry) => entry.command === 'monosmith.doctor')
    expect(doctor?.description).toContain('2 tasks failed')
  })
})
