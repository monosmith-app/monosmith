import { describe, expect, it } from 'vitest'
import { detectRootProject } from './detect-root-project.policy.ts'

const probe = (files: Record<string, string>) => ({
  exists: (path: string) => Object.hasOwn(files, path),
  read:   (path: string) => files[path],
})

describe('detectRootProject', () => {
  it('reads a Node app by its start script', () => {
    expect(detectRootProject('/r/solo', probe({ '/r/solo/package.json': '{"name":"@acme/WebApp","scripts":{"start":"node ."}}' })))
      .toEqual({ manifest: 'package.json', ecosystem: 'node', name: 'web-app', app: true })
  })

  it('reads a Node library, and falls back to the folder name', () => {
    expect(detectRootProject('/r/My Kit', probe({ '/r/My Kit/package.json': '{"main":"index.js"}' })))
      .toEqual({ manifest: 'package.json', ecosystem: 'node', name: 'my-kit', app: false })
  })

  it('reads a Go module, a program when main.go is there', () => {
    expect(detectRootProject('/r/x', probe({ '/r/x/go.mod': 'module github.com/acme/Billing-Tool\n', '/r/x/main.go': '' })))
      .toEqual({ manifest: 'go.mod', ecosystem: 'go', name: 'billing-tool', app: true })
  })

  it('reads a Python project and a Rust crate', () => {
    expect(detectRootProject('/r/x', probe({ '/r/x/pyproject.toml': '[project]\nname = "data_tools"\n' })))
      .toEqual({ manifest: 'pyproject.toml', ecosystem: 'python', name: 'data-tools', app: false })
    expect(detectRootProject('/r/x', probe({ '/r/x/Cargo.toml': '[package]\nname = "rr"\n', '/r/x/src/main.rs': '' }))?.app).toBe(true)
  })

  it('leaves workspaces and ambiguous roots to smith adopt', () => {
    expect(detectRootProject('/r/x', probe({ '/r/x/package.json': '{"workspaces":["a/*"]}' }))).toBeUndefined()
    expect(detectRootProject('/r/x', probe({ '/r/x/go.mod': 'module x\n', '/r/x/go.work': '' }))).toBeUndefined()
    expect(detectRootProject('/r/x', probe({ '/r/x/Cargo.toml': '[workspace]\n' }))).toBeUndefined()
    expect(detectRootProject('/r/x', probe({ '/r/x/go.mod': 'module x\n', '/r/x/package.json': '{}' }))).toBeUndefined()
    expect(detectRootProject('/r/x', probe({ '/r/x/package.json': '{}', '/r/x/pom.xml': '' }))).toBeUndefined()
  })

  it('never offers a folder that has a smith.toml', () => {
    expect(detectRootProject('/r/x', probe({ '/r/x/package.json': '{}', '/r/x/smith.toml': 'broken =' }))).toBeUndefined()
  })

  it('finds nothing in a folder with no manifest or with broken JSON', () => {
    expect(detectRootProject('/r/x', probe({}))).toBeUndefined()
    expect(detectRootProject('/r/x', probe({ '/r/x/package.json': '{oops' }))).toBeUndefined()
  })
})
