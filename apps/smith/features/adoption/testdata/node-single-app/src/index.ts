import { z } from 'zod'

export const Order = z.object({ id: z.string(), total: z.number() })

export function total(orders: z.infer<typeof Order>[]): number {
  return orders.reduce((sum, o) => sum + o.total, 0)
}

console.log(total([{ id: 'a', total: 2 }]))
