import { describe, expect, it } from 'vitest'
import { deriveProjectName } from './derive-project-name.mapper.ts'

describe('deriveProjectName', () => {
  it.each([
    ['@acme/WebApp', 'x', 'web-app'],
    ['my_package', 'x', 'my-package'],
    ['github.com/acme/Billing-Tool', 'x', 'billing-tool'],
    ['github.com/acme/billing/v2', 'x', 'billing'],
    ['Data.Tools', 'x', 'data-tools'],
    ['', 'My Repo', 'my-repo'],
    ['123', 'api', 'api'],
    ['@@@', '', 'project'],
  ])('turns %j (fallback %j) into %j', (raw, fallback, expected) => {
    expect(deriveProjectName(raw, fallback)).toBe(expected)
  })
})
