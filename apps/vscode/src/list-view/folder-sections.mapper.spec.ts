import { describe, expect, it } from 'vitest'
import { groupByFolder, readFolderData, tagWithFolder } from './folder-sections.mapper.ts'

const row = { id: 'x', label: 'x' }

describe('groupByFolder', () => {
  it('leaves the rows of a single folder alone', () => {
    expect(groupByFolder([{ root: '/a', name: 'a', nodes: [row] }])).toEqual([row])
  })

  it('gives each of several folders a row that holds its rows', () => {
    const grouped = groupByFolder([
      { root: '/a', name: 'a', nodes: [row] },
      { root: '/b', name: 'b', nodes: [] },
    ])
    expect(grouped).toMatchObject([
      { id: 'folder:/a', label: 'a', expanded: true, children: [row] },
      { id: 'folder:/b', label: 'b', children: [] },
    ])
  })

  it('has no rows for no folders', () => {
    expect(groupByFolder([])).toEqual([])
  })
})

describe('tagWithFolder', () => {
  it('tells a row of data which folder it is from, at any depth', () => {
    const tagged = tagWithFolder([{ id: 'a', label: 'a', children: [{ id: 'b', label: 'b', data: 'eslint' }, { id: 'c', label: 'c' }] }], '/repo')
    expect(readFolderData(tagged[0]?.children?.[0]?.data)).toEqual({ root: '/repo', data: 'eslint' })
    expect(tagged[0]?.children?.[1]?.data).toBeUndefined()
  })

  it('reads nothing from data that was not tagged', () => {
    expect(readFolderData('eslint')).toBeUndefined()
    expect(readFolderData(undefined)).toBeUndefined()
  })
})
