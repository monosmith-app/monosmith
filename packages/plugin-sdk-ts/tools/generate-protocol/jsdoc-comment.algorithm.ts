const WIDTH = 100

/**
 * A description as a JSDoc block at an indentation, wrapped at 100 columns,
 * paragraphs kept. Returns `''` for no text. A comment terminator inside the
 * text is escaped, so it cannot end the comment early.
 */
export function jsdocComment (text: string | undefined, indent = ''): string {
  if (text === undefined || text.trim() === '') return ''
  const paragraphs = text.replaceAll('*/', String.raw`*\/`).split('\n\n')
  const lines = paragraphs.flatMap((paragraph, i) => [
    ...(i > 0 ? [''] : []),
    ...wrap(paragraph.replaceAll('\n', ' '), WIDTH - indent.length - 3),
  ])
  if (lines.length === 1) return `${indent}/** ${lines[0]} */\n`
  const body = lines.map(line => `${indent} *${line === '' ? '' : ` ${line}`}\n`)

  return `${indent}/**\n${body.join('')}${indent} */\n`
}

function wrap (text: string, width: number): string[] {
  const lines: string[] = []
  let line = ''
  for (const word of text.split(' ')) {
    if (line !== '' && line.length + 1 + word.length > width) {
      lines.push(line)
      line = word
    } else {
      line = line === '' ? word : `${line} ${word}`
    }
  }
  if (line !== '') lines.push(line)

  return lines
}
