import { componentName } from './component-name.algorithm.cjs'

describe('componentName', () => {
  it.each([
    ['logo', 'Logo'],
    ['my-icon', 'MyIcon'],
    ['arrow_left', 'ArrowLeft'],
    ['icon.small', 'IconSmall'],
    ['already Spaced name', 'AlreadySpacedName'],
    ['trailing--', 'Trailing'],
    ['24-hours', '24Hours'],
  ])('names %s %s', (input, expected) => {
    expect(componentName(input)).toBe(expected)
  })
})
