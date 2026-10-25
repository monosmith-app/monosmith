import { describe, expect, it } from 'vitest'
import { stackPlugins } from './stack-plugins.algorithm.ts'

describe('stackPlugins', () => {
  it('lists toolchain, structure and deploy plugins in the order they apply', () => {
    expect(stackPlugins({ toolchain: ['typescript', 'eslint'], structure: ['vertical-slices'], deploy: ['docker'] })).toEqual(['typescript', 'eslint', 'vertical-slices', 'docker'])
    expect(stackPlugins({})).toEqual([])
  })
})
