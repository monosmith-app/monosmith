import { describe, expect, it } from 'vitest'
import { isManifest } from './manifest-match.algorithm.ts'

describe('isManifest', () => {
  it('matches a name exactly and a pattern as Go path.Match does', () => {
    expect(isManifest(['package.json', '*.csproj'], 'package.json')).toBe(true)
    expect(isManifest(['package.json', '*.csproj'], 'App.csproj')).toBe(true)
    expect(isManifest(['package.json', '*.csproj'], 'App.fsproj')).toBe(false)
    expect(isManifest(['pom.xml'], 'pom.xml.bak')).toBe(false)
    expect(isManifest(['?.json'], 'a.json')).toBe(true)
    expect(isManifest(['?.json'], 'ab.json')).toBe(false)
    expect(isManifest(['[a-c]file.txt'], 'bfile.txt')).toBe(true)
    expect(isManifest(['[^a-c]file.txt'], 'bfile.txt')).toBe(false)
    expect(isManifest(['*'], 'anything')).toBe(true)
    expect(isManifest([String.raw`a\*`], 'a*')).toBe(true)
  })

  it('treats a malformed pattern as matching nothing', () => {
    expect(isManifest(['[a-'], 'a')).toBe(false)
    expect(isManifest(['a\\'], 'a')).toBe(false)
    expect(isManifest([], 'a')).toBe(false)
  })
})
