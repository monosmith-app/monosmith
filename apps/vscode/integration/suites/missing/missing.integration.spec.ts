import * as assert from 'node:assert/strict'
import { getExtensionState, waitUntil } from '../../extension-handle/index.ts'

describe('MonoSmith with no smith to be found', () => {
  it('says smith is missing instead of staying empty', async () => {
    const state = await getExtensionState()
    await waitUntil('the view to say smith is missing', () => state.viewState() === 'missing')
    assert.deepEqual(state.projects(), [])
    assert.equal(state.isRunning(), false)
  })
})
