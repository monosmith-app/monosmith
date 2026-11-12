package astro

import "github.com/monosmith-app/monosmith/libs/plugin-node/toolchainversions"

// Versions the starter pins; bump them deliberately.
var devDependencies = map[string]string{
	"@astrojs/check": "^0.9.10",
	"typescript":     toolchainversions.TypeScript,
}

// templates are the site: Starlight with a splash page, a guide and a
// reference page, each section's sidebar generated from its folder.
var templates = map[string]string{
	// build checks first, so astro check and astro build never write the
	// generated .astro/ at the same time.
	"[[.Dir]]/package.json": `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "smith": { "kind": "app"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
  "scripts": {
    "build": "astro check && astro build",
    "start": "astro dev"
  },
  "dependencies": {
    "@astrojs/starlight": "^0.42.6",
    "astro": "^7.4.0"
  }
}
`,
	// Astro's generated types, never a source.
	"[[.Dir]]/.gitignore": `.astro/
`,
	"[[.Dir]]/eslint.config.mjs": `// The workspace's rules, without Astro's generated .astro/.
import workspace from '[[.Up]]eslint.config.mjs'

export default [
  { ignores: ['.astro/'] },
  ...workspace,
]
`,
	"[[.Dir]]/astro.config.mjs": `import starlight from '@astrojs/starlight'
import { defineConfig } from 'astro/config'

// The site builds into the workspace's dist/ (smith), never dist/ in the app.
export default defineConfig({
  outDir:       '[[.Build]]',
  // Bundle js-yaml from Starlight's own copy: externalised, it resolves
  // from the app to the workspace's hoisted js-yaml, a major version apart.
  vite:         { resolve: { noExternal: ['js-yaml'] } },
  integrations: [
    starlight({
      title:   '[[.Name]]',
      sidebar: [
        { label: 'Guides', items: [{ autogenerate: { directory: 'guides' } }] },
        { label: 'Reference', items: [{ autogenerate: { directory: 'reference' } }] },
      ],
    }),
  ],
})
`,
	"[[.Dir]]/src/content.config.ts": `import { defineCollection } from 'astro:content'
import { docsLoader } from '@astrojs/starlight/loaders'
import { docsSchema } from '@astrojs/starlight/schema'

export const collections = {
  docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
}
`,
	"[[.Dir]]/tsconfig.json": `{
  "extends": "astro/tsconfigs/strict",
  "include": [".astro/types.d.ts", "**/*"]
}
`,
	"[[.Dir]]/src/content/docs/index.mdx": `---
title: [[.Name]]
description: The documentation site.
template: splash
hero:
  tagline: Everything about the project, in one place.
  actions:
    - text: Get started
      link: guides/getting-started/
      icon: right-arrow
---
`,
	"[[.Dir]]/src/content/docs/guides/getting-started.md": `---
title: Getting started
description: The first steps.
---

Write guides as Markdown under ` + "`src/content/docs/guides`" + `.
`,
	"[[.Dir]]/src/content/docs/reference/overview.md": `---
title: Overview
description: The reference.
---

Write reference pages as Markdown under ` + "`src/content/docs/reference`" + `.
`,
}
