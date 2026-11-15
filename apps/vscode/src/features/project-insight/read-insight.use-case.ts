import type { SmithConnection } from '../smith-client/index.ts'
import { toExplainDocument } from './explain-document.mapper.ts'
import { toGraphDocument } from './graph-document.mapper.ts'
import type { ExplainedTask, GraphAnswer } from './insight-answer.contract.ts'

/** How a project's tasks resolved, as a Markdown page. */
export async function readExplain (connection: SmithConnection, project: string): Promise<string> {
  const tasks = await connection.request<ExplainedTask[]>('explain', { target: project })

  return toExplainDocument(project, Array.isArray(tasks) ? tasks : [])
}

/** What a project uses and what uses it, as a Markdown page. */
export async function readGraph (connection: SmithConnection, project: string): Promise<string> {
  const answer = await connection.request<GraphAnswer>('graph', { focus: [project] })

  return toGraphDocument(project, answer)
}
