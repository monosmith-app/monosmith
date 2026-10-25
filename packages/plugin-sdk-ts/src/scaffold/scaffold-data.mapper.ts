import type { GenerateParams } from '../protocol/index.ts'

/**
 * What a project template can say, from the `generate` params:
 *
 * | Field | Value |
 * |---|---|
 * | `Dir` | the project's directory, workspace-relative |
 * | `Name` | the project's name |
 * | `Kind` | `app`, `lib` or `package` |
 * | `Stack` | the stack to record in the manifest (may be empty) |
 * | `Workspace` | the workspace's name |
 * | `Scope` | the npm scope, with its `@` (may be empty) |
 * | `ModulePrefix` | the Go module path prefix (may be empty) |
 */
export function scaffoldData (params: GenerateParams): Record<string, string> {
  return {
    Dir:          params.directory,
    Name:         params.name,
    Kind:         params.kind,
    Stack:        params.stack ?? '',
    Workspace:    params.workspace.name,
    Scope:        params.workspace.scope ?? '',
    ModulePrefix: params.workspace.modulePrefix ?? '',
  }
}
