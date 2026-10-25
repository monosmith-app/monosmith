import type { GenerateParams, GenerateResult } from '../protocol/index.ts'
import { writeTemplates } from '../templating/index.ts'
import { scaffoldData } from './scaffold-data.mapper.ts'

export interface ScaffoldOptions {
  /** The workspace root from `initialize`: the project's directory is relative to it. */
  workspaceRoot: string
  /** The `generate` params. */
  params:        GenerateParams
  /** Target path to template text, with `[[ ]]` actions in both (`'[[.Dir]]/widget.json'`). */
  files:         Record<string, string>
  /** More fields for the templates, next to the ones {@link scaffoldData} lists. They win over them. */
  data?:         Record<string, unknown>
  /** Notes printed to the user after generation. */
  notes?:        string[]
}

/**
 * The answer to `generate` for a scaffold: renders the templates into the new
 * project, never overwriting a file, and reports what it wrote. A language
 * plugin that scaffolds must record `params.stack` in the manifest it writes
 * (`[[if .Stack]]` in a template), so discovery reports it back.
 */
export async function scaffoldProject (options: ScaffoldOptions): Promise<GenerateResult> {
  const data = { ...scaffoldData(options.params), ...options.data }
  const files = await writeTemplates(options.workspaceRoot, options.files, data)

  return options.notes === undefined ? { files } : { files, notes: options.notes }
}
