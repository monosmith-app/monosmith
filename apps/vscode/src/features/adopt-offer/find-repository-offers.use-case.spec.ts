import { describe, expect, it } from 'vitest'
import { findRepositoryOffers } from './find-repository-offers.use-case.ts'

const probe = { exists: (path: string) => path.endsWith('nx.json'), read: () => undefined }

describe('findRepositoryOffers', () => {
  it('offers the folders that look like a repository', () => {
    const offers = findRepositoryOffers([{ root: '/a', name: 'a' }], new Set(), probe)
    expect(offers).toEqual([{ root: '/a', name: 'a', kind: 'nx', marker: 'nx.json' }])
  })

  it('leaves out folders that already run smith', () => {
    expect(findRepositoryOffers([{ root: '/a', name: 'a' }], new Set(['/a']), probe)).toEqual([])
  })

  it('offers nothing for a folder with no marker', () => {
    expect(findRepositoryOffers([{ root: '/a', name: 'a' }], new Set(), { exists: () => false, read: () => undefined })).toEqual([])
  })
})
