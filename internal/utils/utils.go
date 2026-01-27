package utils

import (
	"math"
)

// Math helper functions for concentrated liquidity calculations
func TickToSqrtPrice(tick int) float64 {
	return math.Sqrt(math.Pow(1.0001, float64(tick)))
}

func GetAmounts(liquidity float64, tickLower, tickUpper, currentTick int) (float64, float64) {
	currentPrice := TickToSqrtPrice(currentTick)
	lowerPrice := TickToSqrtPrice(tickLower)
	upperPrice := TickToSqrtPrice(tickUpper)

	var amount0, amount1 float64

	if currentPrice < lowerPrice {
		amount1 = 0
		amount0 = liquidity * (1/lowerPrice - 1/upperPrice)
	} else if lowerPrice <= currentPrice && currentPrice <= upperPrice {
		amount1 = liquidity * (currentPrice - lowerPrice)
		amount0 = liquidity * (1/currentPrice - 1/upperPrice)
	} else {
		amount1 = liquidity * (upperPrice - lowerPrice)
		amount0 = 0
	}

	return amount0, amount1
}
