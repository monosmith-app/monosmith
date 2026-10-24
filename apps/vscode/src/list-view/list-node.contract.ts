/** One row of a MonoSmith list view (health, dependencies, release). */
export interface ListNode {
  id:           string
  label:        string
  description?: string
  tooltip?:     string
  /** A codicon id, such as `error` or `arrow-up`. */
  icon?:        string
  /** A theme color id, such as `charts.red`. */
  color?:       string
  /** What menus see as `viewItem`. */
  context?:     string
  /** Shown open; a row with children is closed otherwise. */
  expanded?:    boolean
  /** What a command given the row needs, such as a package name. */
  data?:        string
  children?:    ListNode[]
}

/** What a list view shows once it has loaded. */
export interface ListResult {
  nodes:    ListNode[]
  /** Said above the rows, such as why there are none. */
  message?: string
  /** Said beside the view's title. */
  summary?: string
}
