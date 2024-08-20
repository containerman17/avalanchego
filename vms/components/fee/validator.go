// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package fee

import (
	"fmt"
)

type ValidatorState struct {
	Current                  Gas
	Target                   Gas
	Capacity                 Gas
	Excess                   Gas
	MinFee                   GasPrice
	ExcessConversionConstant Gas
}

func (v ValidatorState) CurrentFeeRate() GasPrice {
	return v.MinFee.MulExp(v.Excess, v.ExcessConversionConstant)
}

func (v ValidatorState) CalculateContinuousFee(seconds uint64) uint64 {
	if v.Current == v.Target {
		return uint64(v.MinFee.MulExp(v.Excess, v.ExcessConversionConstant)) * seconds
	}

	var totalFee uint64
	if v.Current < v.Target {
		secondsTillExcessIsZero := uint64(v.Excess / (v.Target - v.Current))

		if secondsTillExcessIsZero < seconds {
			totalFee += uint64(v.MinFee) * (seconds - secondsTillExcessIsZero)
			seconds = secondsTillExcessIsZero
		}
	}

	x := v.Excess
	for i := uint64(0); i < seconds; i++ {
		if v.Current < v.Target {
			x = x.SubPerSecond(v.Target-v.Current, 1)
		} else {
			x = x.AddPerSecond(v.Current-v.Target, 1)
		}

		totalFee += uint64(v.MinFee.MulExp(x, v.ExcessConversionConstant))
	}

	return totalFee
}

// Returns the first number n where CalculateContinuousFee(n) >= balance
func (v ValidatorState) CalculateTimeTillContinuousFee(balance uint64) (uint64, uint64) {
	// Lower bound can be derived from [MinFee].
	n := balance / uint64(v.MinFee)
	interval := n

	numIters := 0
	for {
		fmt.Printf("n=%d", n)
		feeAtN := v.CalculateContinuousFee(n)
		feeBeforeN := v.CalculateContinuousFee(n - 1)
		if feeAtN == balance {
			return n, uint64(numIters)
		}

		if feeAtN > balance && feeBeforeN < balance {
			return n, uint64(numIters)
		}

		if feeAtN > balance {
			if interval > 1 {
				interval /= 2
			}
			n -= interval
		}

		if feeAtN < balance {
			n += interval
		}

		numIters++
	}
}
