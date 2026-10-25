import { readFileSync } from 'node:fs'
import { join, posix } from 'node:path'
import { fileURLToPath } from 'node:url'
import { JsonDocument, buildTask, scaffoldProject, task } from '@monosmith/plugin-sdk'
import type { Finding, Project, ProjectKind, SmithPlugin } from '@monosmith/plugin-sdk'
import { MANIFEST_FILE, readManifest, setField } from './manifest.ts'

/**
 * The command the tasks run: this program. A plugin that wraps a real tool
 * would name the tool here instead.
 */
const SELF = ['node', fileURLToPath(new URL('plugin.ts', import.meta.url))]

/**
 * The plugin. Every method it has beyond describe is a capability the SDK
 * advertises: discover, generate, writeVersion and doctor.
 */
export function createWidgetPlugin (): SmithPlugin {
  let workspaceRoot = ''
  const manifestPath = (relative: string): string => join(workspaceRoot, ...relative.split('/'))

  return {
    describe: (params) => {
      workspaceRoot = params.workspaceRoot

      return {
        plugin:    { name: 'widget', version: '0.1.0' },
        kind:      'language',
        // A language plugin that also writes new projects plays the scaffold level too;
        // it does not need to be chosen as a toolchain.
        roles:     ['language', 'scaffold'],
        languages: ['widget'],
        manifests: [MANIFEST_FILE],
      }
    },

    // One project per widget.json that is ours.
    discover: ({ manifests }) => {
      const projects: Project[] = []
      for (const relative of manifests) { // workspace-relative, forward slashes
        const manifest = readManifest(manifestPath(relative))
        if (manifest.name === '') continue // not a widget manifest: leave it out
        const root = posix.dirname(relative)
        projects.push({
          name:        manifest.name,
          root,
          language:    'widget',
          kind:        (manifest.kind === '' ? 'lib' : manifest.kind) as ProjectKind,
          version:     manifest.version,
          description: manifest.description,
          stack:       manifest.stack,
          exports:     [manifest.name],
          imports:     manifest.uses, // the host keeps the ones another project exports
          tasks:       {
            build: buildTask(root, outputDirectory => [...SELF, 'build', outputDirectory], {
              inputs:      [MANIFEST_FILE],
              description: 'write widget.txt into the build directory',
            }),
            test: task([...SELF, 'check'], {
              inputs:      [MANIFEST_FILE],
              cache:       true,
              description: 'check the manifest: a name and an x.y.z version',
            }),
          },
        })
      }

      return { projects }
    },

    // A new project's starter files. The templates use [[ ]], so a body can contain {{ }} untouched.
    generate: async (params) => await scaffoldProject({
      workspaceRoot,
      params,
      files: {
        '[[.Dir]]/widget.json': `{
  "name": "[[.Name]]",
  "version": "0.0.0",
  "kind": "[[.Kind]]",
  "description": "A widget of the [[.Workspace]] workspace.",
  "license": "MIT"[[if .Stack]],
  "stack": "[[.Stack]]"[[end]]
}
`,
        '[[.Dir]]/README.md': '# [[.Name]]\n\nA widget. Edit widget.json.\n',
      },
    }),

    // Record the release version.
    writeVersion: ({ project, version, dryRun }) => {
      const file = posix.join(project.root, MANIFEST_FILE)
      if (dryRun !== true) setField(manifestPath(file), 'version', version)

      return { files: [file] }
    },

    // Every failed finding names its remedy, and a fix when the plugin can write one.
    doctor: ({ projects }) => ({
      findings: projects.map((project): Finding => {
        const file = posix.join(project.root, MANIFEST_FILE)
        const manifest = readManifest(manifestPath(file))
        const check = `${project.name} names a license`
        if (manifest.license !== '') return { check, ok: true }
        const document = JsonDocument.parse(readFileSync(manifestPath(file), 'utf8'))
        document.setString('license', 'MIT')

        return {
          check,
          ok:     false,
          detail: `${file} has no license field`,
          remedy: `add "license": "MIT" (or your license) to ${file}`,
          // Not safe: it picks a license for the team, so smith asks first.
          fix:    {
            description: 'add an MIT license to widget.json',
            safe:        false,
            applies_to:  [project.name],
            files:       [{ path: file, content: document.render() }],
          },
        }
      }),
    }),
  }
}
