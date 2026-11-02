import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { parseKinds } from './add-options.mapper.ts'
import { fitOutcomes, outcomeRows, parseOutcomes } from './add-outcomes.mapper.ts'

/** Real output of `smith compose --json` and `smith kinds --json` (2.0.0-beta.23), not invented JSON. */
const compose = readFileSync(new URL('fixtures/compose-report.fixture.json', import.meta.url), 'utf8')
const kinds = readFileSync(new URL('fixtures/kinds.fixture.json', import.meta.url), 'utf8')

describe('the first question of Add, read from the real output of smith', () => {
  const fitted = fitOutcomes(parseOutcomes(compose).outcomes, parseKinds(kinds))
  const found = (query: string): string[] => outcomeRows(fitted, query).flatMap((row) => row.kind === 'outcome' ? row.outcome.id : [])

  it('offers all six outcomes, the website and the desktop app included', () => {
    expect(fitted.map((outcome) => outcome.id)).toEqual(['web', 'api', 'function', 'desktop', 'library', 'package'])
  })

  it('names a desktop shell and a cloud host with the title of the catalogue, the one name of every step', () => {
    const { layers } = parseOutcomes(compose)

    expect(layers.shell).toBe('Desktop shell')
    expect(layers.host).toBe('Cloud host')
  })

  it('finds the website for react and the desktop app for electron', () => {
    expect(found('react')[0]).toBe('web')
    expect(found('electron')[0]).toBe('desktop')
  })
})
