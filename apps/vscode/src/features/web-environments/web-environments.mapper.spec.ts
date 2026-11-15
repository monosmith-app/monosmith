import { describe, expect, it } from 'vitest'
import { checkEnvironmentNames, defaultEnvironment, defaultFirst, describeEnvironments, environmentArguments, isWebApp, parseEnvironmentNames, startingEnvironments } from './web-environments.mapper.ts'

const web = { role: 'app', language: 'node', with: ['react'], stack: undefined }

describe('isWebApp', () => {
  it('is a Vite app: a node app built from react, vue, svelte or html, or from a stack of them', () => {
    expect(isWebApp(web)).toBe(true)
    expect(isWebApp({ ...web, with: ['html', 'vitest'] })).toBe(true)
    expect(isWebApp({ ...web, with: [], stack: 'react-tauri' })).toBe(true)
  })

  it('is not an API, a library or another language', () => {
    expect(isWebApp({ ...web, with: ['express'] })).toBe(false)
    expect(isWebApp({ ...web, role: 'lib' })).toBe(false)
    expect(isWebApp({ ...web, language: 'go' })).toBe(false)
  })
})

describe('environment names', () => {
  it('reads commas, spaces and semicolons', () => {
    expect(parseEnvironmentNames('lcl, dev  uat;prd,')).toEqual(['lcl', 'dev', 'uat', 'prd'])
  })

  it('accepts lowercase words', () => {
    expect(checkEnvironmentNames('lcl, dev, uat, prd')).toBeUndefined()
  })

  it('names the fix for each mistake', () => {
    expect(checkEnvironmentNames('')).toMatch(/at least one/)
    expect(checkEnvironmentNames('UAT')).toMatch(/lowercase word/)
    expect(checkEnvironmentNames('production')).toMatch(/already give it a meaning/)
    expect(checkEnvironmentNames('dev, dev')).toMatch(/twice/)
  })

  it('builds prd by default, else the last one', () => {
    expect(defaultEnvironment(['lcl', 'dev', 'uat', 'prd'])).toBe('prd')
    expect(defaultEnvironment(['dev', 'stage'])).toBe('stage')
  })
})

describe('the flags of the environments', () => {
  it('says every answer, so smith asks nothing', () => {
    const environments = { names: ['lcl', 'prd'], default: 'prd', output: 'workspace' as const, envFiles: 'commit' as const }
    expect(environmentArguments(environments)).toEqual(['--environments', 'lcl,prd', '--default-environment', 'prd', '--environment-output', 'workspace', '--env-files', 'commit'])
  })
})

describe('the single environments question', () => {
  it('starts on prd, else the last name, never on lcl', () => {
    expect(startingEnvironments([])).toEqual({ names: ['lcl', 'dev', 'uat', 'prd'], default: 'prd', output: 'workspace', envFiles: 'commit' })
    expect(startingEnvironments(['lcl', 'qa']).default).toBe('qa')
  })

  it('lists the default first so Enter takes it', () => {
    expect(defaultFirst(['lcl', 'dev', 'prd'], 'prd')).toEqual(['prd', 'lcl', 'dev'])
  })

  it('says every choice in one sentence', () => {
    expect(describeEnvironments(startingEnvironments([]))).toBe(
      'Environments: lcl, dev, uat, prd; smith run build builds prd; builds go to dist/apps/<name>/<env>; .env files are created and kept by git',
    )
    expect(describeEnvironments({ ...startingEnvironments([]), output: 'project', envFiles: 'example' })).toContain('dist/<env> inside the app; .env files are created and ignored by git')
  })
})
