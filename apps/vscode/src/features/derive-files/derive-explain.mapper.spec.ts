import { describe, expect, it } from 'vitest'
import { toExplainDocument } from './derive-explain.mapper.ts'

describe('toExplainDocument', () => {
  it('puts the text of smith in a diff fence under the file name', () => {
    const page = toExplainDocument('a/tsconfig.json', 'a/tsconfig.json\n  owner     smith\n--- x\n+++ y\n')
    expect(page).toContain('# Derived file: a/tsconfig.json')
    expect(page).toContain('```diff\na/tsconfig.json\n  owner     smith\n--- x\n+++ y\n```')
  })
})
