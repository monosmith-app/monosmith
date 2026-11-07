package marks

// Passed marks a step, check or file that succeeded.
const Passed = "✓"

// Failed marks a step, check or file that failed.
const Failed = "×"

// Warning marks a result that needs a look but does not stop the run.
const Warning = "▲"

// Skipped marks a step that did not run.
const Skipped = "»"

// Cancelled marks a step the person stopped. It is neutral: the step did not
// fail, it was interrupted.
const Cancelled = "⊘"

// Cached marks a step whose result came from the cache.
const Cached = "≈"
