import { describe, expect, it } from 'vitest'
import { isPreview, readPreview, readWrite, sayProblem } from './install-outcome.mapper.ts'

const PREVIEW = '● Add .vscode/mcp.json\n--- a/.vscode/mcp.json\n+++ b/.vscode/mcp.json\n+{}\n\n● Add .mcp.json\n--- a/.mcp.json\n+++ b/.mcp.json\n+{}\n'
const NO_TERMINAL = '✖ nothing was written: there is no terminal to ask on\n  → run it again with --yes to write the files above\n'

describe('readPreview', () => {
  it('reads the files from the diffs smith prints before it stops for want of a terminal', () => {
    const preview = readPreview({ code: 1, stdout: PREVIEW, stderr: NO_TERMINAL })
    expect(isPreview(preview) && preview.actions).toEqual(['Add .vscode/mcp.json', 'Add .mcp.json'])
    expect(isPreview(preview) && preview.diff).toContain('--- a/.mcp.json')
  })

  it('reads current files as nothing to do', () => {
    expect(readPreview({ code: 0, stdout: '✓ The MonoSmith agent is already installed (3 files)\n', stderr: '' })).toEqual({ kind: 'current' })
  })

  it('refuses with smith’s words when a file was edited', () => {
    const stderr = '✖ .mcp.json was changed after smith wrote it, so smith left everything as it is\n  → keep your version: nothing to do (smith doctor accepts it)\n'
    const outcome = readPreview({ code: 1, stdout: '● .mcp.json differs from what smith writes\n', stderr })
    expect(outcome).toEqual({
      kind:    'refused',
      message: '.mcp.json was changed after smith wrote it, so smith left everything as it is Try: keep your version: nothing to do (smith doctor accepts it)',
    })
  })

  it('refuses a file with comments, and a smith that did not start', () => {
    expect(readPreview({ code: 1, stdout: '', stderr: '✖ .vscode/mcp.json is not plain JSON (comments?)\n  → add the smith server by hand\n' })).toMatchObject({ kind: 'refused' })
    expect(readPreview({ code: 1, stdout: '', stderr: '' })).toEqual({ kind: 'refused', message: 'smith stopped with exit code 1' })
  })
})

describe('readWrite', () => {
  it('lists the files written', () => {
    expect(readWrite({ code: 0, stdout: '✓ Wrote .mcp.json\n✓ Wrote .vscode/mcp.json\n→ VS Code: open .vscode/mcp.json\n', stderr: '' }))
      .toEqual({ kind: 'written', lines: ['Wrote .mcp.json', 'Wrote .vscode/mcp.json'] })
  })

  it('says current when nothing was written', () => {
    expect(readWrite({ code: 0, stdout: '', stderr: '' })).toEqual({ kind: 'current' })
  })

  it('refuses on a failure', () => {
    expect(readWrite({ code: 1, stdout: '', stderr: '✖ no smith.toml here\n' })).toEqual({ kind: 'refused', message: 'no smith.toml here' })
  })
})

describe('sayProblem', () => {
  it('drops the marks and keeps each way out', () => {
    expect(sayProblem('✖ it broke\n  → do this\n  → or that\n')).toBe('it broke Try: do this Try: or that')
  })
})
