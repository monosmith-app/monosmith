// Package orders adds up orders.
package orders

// Total adds the amounts.
func Total(amounts []float64) float64 {
	sum := 0.0
	for _, a := range amounts {
		sum += a
	}
	return sum
}
