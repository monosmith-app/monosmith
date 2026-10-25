// Turning a schema description into the XML documentation comment of a C# member.

const escapeXml = (text) => text.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;')

// Backticked words become <c> elements, as the descriptions use them for code.
const inline = (text) => escapeXml(text.replace(/\s*\n\s*/g, ' ').trim()).replace(/`([^`]+)`/g, '<c>$1</c>')

/**
 * The documentation comment of a member: the first paragraph of the description is the summary, the
 * others are remarks. `extra` is a sentence added to the summary (for a member the schema says nothing about).
 */
export function docComment(description, indent = '', fallback = '') {
  const text = (description ?? '').trim()
  const paragraphs = text === '' ? [] : text.split(/\n\s*\n/)
  const [first, ...rest] = paragraphs.length > 0 ? paragraphs : [fallback]
  if (first === undefined || first === '') return ''
  const lines = [`${indent}/// <summary>${inline(first)}</summary>`]
  if (rest.length > 0) {
    lines.push(`${indent}/// <remarks>`)
    for (const paragraph of rest) lines.push(`${indent}/// <para>${inline(paragraph)}</para>`)
    lines.push(`${indent}/// </remarks>`)
  }
  return lines.join('\n') + '\n'
}

/** A one-line summary for a member without a description. */
export function summaryLine(text, indent = '') {
  return `${indent}/// <summary>${inline(text)}</summary>\n`
}
