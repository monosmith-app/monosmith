import { describe, expect, it } from 'vitest'
import { ROLES, defaultRoles, effectiveRoles, isAdditiveRole, isExclusiveRole, isKnownRole } from './plugin-role.policy.ts'

describe('plugin roles', () => {
  it('lists the roles in the order of a project\'s make-up', () => {
    expect(ROLES).toEqual(['language', 'toolchain', 'scaffold', 'structure', 'deploy'])
    expect(isKnownRole('scaffold')).toBe(true)
    expect(isKnownRole('platform')).toBe(false)
  })

  it('says how many plugins may play one', () => {
    expect(ROLES.filter(isExclusiveRole)).toEqual(['language', 'scaffold'])
    expect(ROLES.filter(isAdditiveRole)).toEqual(['structure', 'deploy'])
  })

  it('defaults the roles of a plugin that names none by its kind', () => {
    expect(defaultRoles('language')).toEqual(['language', 'toolchain', 'scaffold'])
    expect(defaultRoles('convention')).toEqual(['structure'])
    expect(effectiveRoles({ kind: 'language' })).toEqual(['language', 'toolchain', 'scaffold'])
    expect(effectiveRoles({ kind: 'language', roles: [] })).toEqual(['language', 'toolchain', 'scaffold'])
    expect(effectiveRoles({ kind: 'convention', roles: ['deploy'] })).toEqual(['deploy'])
  })
})
