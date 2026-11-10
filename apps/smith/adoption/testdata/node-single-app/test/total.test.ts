import { total } from '../src/index'

test('adds the totals', () => {
  expect(total([{ id: 'a', total: 2 }, { id: 'b', total: 3 }])).toBe(5)
})
