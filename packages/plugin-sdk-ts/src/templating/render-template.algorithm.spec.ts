import { describe, expect, it } from 'vitest'
import { renderTemplate } from './render-template.algorithm.ts'
import { TemplateError } from './template.error.ts'

const data = { Dir: 'apps/web', Name: 'web', Kind: 'app', Stack: '', Port: 8080, Tags: ['a', 'b'], Workspace: { Name: 'acme', Scope: '@acme' }, Empty: [] as string[] }

describe('renderTemplate', () => {
  it('prints fields, dotted fields and literals', () => {
    expect(renderTemplate('[[.Name]] in [[.Dir]] on [[.Port]] of [[.Workspace.Scope]] [[.Tags]] [[ "[[" ]] [[ 3 ]] [[ true ]]', data)).toBe('web in apps/web on 8080 of @acme [a b] [[ 3 true')
  })

  it('chooses with if, else if and else', () => {
    const t = '[[if eq .Kind "app"]]app[[else if eq .Kind "lib"]]lib[[else]]other[[end]]'
    expect(renderTemplate(t, data)).toBe('app')
    expect(renderTemplate(t, { ...data, Kind: 'lib' })).toBe('lib')
    expect(renderTemplate(t, { ...data, Kind: 'package' })).toBe('other')
    expect(renderTemplate('[[if .Stack]]stack: [[.Stack]][[end]]|[[if not .Stack]]none[[end]]', data)).toBe('|none')
    expect(renderTemplate('[[if ne .Kind "package"]]x[[end]][[if and .Name .Port]]y[[end]][[if or .Stack .Empty]]z[[end]]', data)).toBe('xy')
  })

  it('writes the optional member of a manifest the way the SDK\'s own templates do', () => {
    const t = '{\n  "license": "MIT"[[if .Stack]],\n  "stack": "[[.Stack]]"[[end]]\n}\n'
    expect(renderTemplate(t, data)).toBe('{\n  "license": "MIT"\n}\n')
    expect(renderTemplate(t, { ...data, Stack: 'widget' })).toBe('{\n  "license": "MIT",\n  "stack": "widget"\n}\n')
  })

  it('narrows with with, loops with range, and reaches the root with $', () => {
    expect(renderTemplate('[[with .Workspace]][[.Name]]/[[$.Name]][[end]]', data)).toBe('acme/web')
    expect(renderTemplate('[[range .Tags]]<[[.]]>[[end]]', data)).toBe('<a><b>')
    expect(renderTemplate('[[range .Empty]]x[[else]]none[[end]]', data)).toBe('none')
  })

  it('calls the functions of text/template', () => {
    expect(renderTemplate('[[len .Tags]] [[len .Name]] [[index .Tags 1]] [[lt .Port 9000]] [[ge .Name "web"]] [[printf "%s:%d" .Name .Port]] [[print "a" 1 2 "b"]] [[.Name | printf "%q"]]', data)).toBe('2 3 b true true web:8080 a1 2b "web"')
  })

  it('trims whitespace with the - markers, and skips comments', () => {
    expect(renderTemplate('a  \n[[- .Name -]]\n  b[[/* a comment */]]', data)).toBe('awebb')
    expect(renderTemplate('x[[ -1 ]]', data)).toBe('x-1')
  })

  it('fails with the name of the template and what is wrong', () => {
    const cases: Array<[string, string]> = [
      ['[[.Missing]]', 'map has no entry for key "Missing"'],
      ['[[.Name.Deeper]]', 'can\'t evaluate field Deeper in string'],
      ['[[if .Name]]x', 'unclosed if: missing [[end]]'],
      ['[[end]]', 'unexpected end'],
      ['[[else]]', 'unexpected else'],
      ['[[.Name', 'unclosed action'],
      ['[[nope .Name]]', 'function "nope" not defined'],
      ['[[eq .Name 3]]', 'incompatible types for comparison'],
      ['[[$x := 1]]', 'variables are not supported'],
      ['[[template "x"]]', 'is not supported'],
      ['[[(.Name]]', 'unclosed ('],
      ['[[.Name .Port]]', 'can\'t give an argument'],
    ]
    for (const [text, message] of cases) {
      expect(() => renderTemplate(text, data, 'main.go'), text).toThrow(new TemplateError('main.go', message).message.slice(0, 12))
      expect(() => renderTemplate(text, data, 'main.go'), text).toThrow(message)
    }
  })

  it('treats nothing as an empty data object only when it has no fields', () => {
    expect(renderTemplate('plain text', undefined)).toBe('plain text')
    expect(() => renderTemplate('[[.X]]', undefined)).toThrow('can\'t evaluate field X in nil')
  })
})
