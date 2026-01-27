package utils

import (
	"math"
	"testing"
)

func TestTickToSqrtPrice(t *testing.T) {
	tests := []struct {
		name     string
		tick     int
		expected float64
	}{
		{
			name:     "tick 0",
			tick:     0,
			expected: 1.0,
		},
		{
			name:     "positive tick",
			tick:     1000,
			expected: 1.0512684683767608,
		},
		{
			name:     "negative tick",
			tick:     -1000,
			expected: 0.9512318024187264,
		},
		{
			name:     "large positive tick",
			tick:     887270,
			expected: 18444206290207213568.0,
		},
		{
			name:     "large negative tick",
			tick:     -887270,
			expected: 5.421756752534544e-20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TickToSqrtPrice(tt.tick)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("TickToSqrtPrice(%d) = %f, expected %f", tt.tick, result, tt.expected)
			}
		})
	}
}

func TestGetAmounts(t *testing.T) {
	tests := []struct {
		name        string
		liquidity   float64
		tickLower   int
		tickUpper   int
		currentTick int
		expected0   float64
		expected1   float64
	}{
		{
			name:        "deafult case",
			liquidity:   3276468520,
			tickLower:   -887270,
			tickUpper:   887270,
			currentTick: 339603,
			expected0:   138.48069913477778,
			expected1:   7.752160430739182e+16,
		},
		{
			name:        "current price below range",
			liquidity:   1000000,
			tickLower:   1000,
			tickUpper:   2000,
			currentTick: 500,
			expected0:   46389.86048594762,
			expected1:   0,
		},
		{
			name:        "current price above range",
			liquidity:   1000000,
			tickLower:   1000,
			tickUpper:   2000,
			currentTick: 2500,
			expected0:   0,
			expected1:   53896.924226459756,
		},
		{
			name:        "current price in range",
			liquidity:   1000000,
			tickLower:   1000,
			tickUpper:   2000,
			currentTick: 1500,
			expected0:   22905.02320845944,
			expected1:   26611.64071932465,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount0, amount1 := GetAmounts(tt.liquidity, tt.tickLower, tt.tickUpper, tt.currentTick)

			tolerance := 1e-10
			if tt.expected0 != 0 {
				if math.Abs((amount0-tt.expected0)/tt.expected0) > tolerance {
					t.Errorf("GetAmounts() amount0 = %f, expected %f", amount0, tt.expected0)
				}
			} else if math.Abs(amount0) > tolerance {
				t.Errorf("GetAmounts() amount0 = %f, expected %f", amount0, tt.expected0)
			}

			if tt.expected1 != 0 {
				if math.Abs((amount1-tt.expected1)/tt.expected1) > tolerance {
					t.Errorf("GetAmounts() amount1 = %f, expected %f", amount1, tt.expected1)
				}
			} else if math.Abs(amount1) > tolerance {
				t.Errorf("GetAmounts() amount1 = %f, expected %f", amount1, tt.expected1)
			}
		})
	}
}
