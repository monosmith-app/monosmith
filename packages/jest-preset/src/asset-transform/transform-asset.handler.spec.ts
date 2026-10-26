import { createRequire } from 'node:module'
import transformer from './transform-asset.handler.cjs'

interface SvgModule {
  __esModule:     boolean
  default:        string
  ReactComponent: ((props: object) => { type: string, props: Record<string, unknown> }) & { name: string }
}

/** Runs the module code the transform returned, with the React of this workspace. */
function evaluate<T> (code: string): T {
  const module = { exports: {} as unknown }
  // eslint-disable-next-line @typescript-eslint/no-implied-eval
  new Function('module', 'require', code)(module, createRequire(__filename))

  return module.exports as T
}

/** Calls the transformer the way Jest does, for a file that is not a script. */
function transformAsset (filename: string): string {
  return (transformer.process('', filename, { config: {} }) as { code: string }).code
}

describe('the asset transform', () => {
  it('turns an SVG into a module with its file name and a React component', () => {
    const exported = evaluate<SvgModule>(transformAsset('/src/assets/my-logo.svg'))

    expect(exported.__esModule).toBe(true)
    expect(exported.default).toBe('my-logo.svg')
    expect(exported.ReactComponent.name).toBe('SvgMyLogo')
    const element = exported.ReactComponent({ width: 4 })
    expect(element.type).toBe('svg')
    expect(element.props).toEqual({ width: 4, children: 'my-logo.svg' })
  })

  it('turns any other file into its file name', () => {
    expect(evaluate<string>(transformAsset('/src/assets/photo.png'))).toBe('photo.png')
    expect(evaluate<string>(transformAsset('/src/fonts/a-font.woff2'))).toBe('a-font.woff2')
  })
})
