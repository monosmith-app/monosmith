package main

import (
	"fmt"

	"example.com/acme/orders-svc/orders"
)

func main() {
	fmt.Println(orders.Total([]float64{2, 3}))
}
