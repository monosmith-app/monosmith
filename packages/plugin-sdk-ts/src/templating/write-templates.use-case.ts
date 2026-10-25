import { access, mkdir, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, posix, resolve, sep } from 'node:path'
import { FileExistsError } from './file-exists.error.ts'
import { renderTemplate } from './render-template.algorithm.ts'
import { byKey } from './template-functions.algorithm.ts'
import { TemplateError } from './template.error.ts'

/**
 * Renders every template in `files` (target path to template text) under
 * `root` with `data`, creating directories as needed. A target path may itself
 * contain template actions (`[[.Dir]]/main.ts`). Returns the written paths,
 * forward-slashed and relative to `root`, in the order of the sorted templates.
 *
 * Nothing is written if any template fails to render or any target exists
 * ({@link FileExistsError}): it never overwrites a file. A target that would
 * land outside `root` is refused.
 */
export async function writeTemplates (root: string, files: Record<string, string>, data?: unknown): Promise<string[]> {
  const rendered: Array<{ path: string, content: string }> = []
  for (const target of Object.keys(files).sort(byKey)) {
    const path = posix.normalize(renderTemplate(target, data, 'path'))
    const full = resolve(root, ...path.split('/'))
    if (path === '..' || path.startsWith('../') || isAbsolute(path) || !full.startsWith(resolve(root) + sep)) {
      throw new TemplateError(path, 'the target is outside the project root')
    }
    if (await exists(full)) throw new FileExistsError(path)
    rendered.push({ path, content: renderTemplate(files[target], data, path) })
  }
  for (const file of rendered) {
    const full = join(root, ...file.path.split('/'))
    await mkdir(dirname(full), { recursive: true, mode: 0o755 })
    await writeFile(full, file.content, { mode: 0o644 })
  }

  return rendered.map(file => file.path)
}

async function exists (path: string): Promise<boolean> {
  try {
    await access(path)

    return true
  } catch {
    return false
  }
}
