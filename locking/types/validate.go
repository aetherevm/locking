// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types

import (
	fmt "fmt"
	"strings"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Validation Errors
	ErrInvalidType       = "invalid parameter type: %T"
	ErrInvalidDenom      = "invalid denom: %s"
	ErrZeroTime          = "time cannot be zero: %s"
	ErrZeroDec           = "dec cannot be zero: %s"
	ErrZeroOrNegativeInt = "int cannot be zero or negative: %s"
	ErrZeroCoin          = "coin cannot be zero: %s"
	ErrZeroDuration      = "duration cannot be zero: %s"
	ErrZeroU32           = "u32 cannot be zero: %d"
)

// ValidateDenom validates if the given parameter is a non-empty string.
// It returns an error if the parameter is not of type string or if it is empty.
func ValidateDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf(ErrInvalidDenom, v)
	}

	return nil
}

// ValidateNonZeroTime validates if the given parameter is a non-zero time.
// It returns an error if the parameter is not of type time.Time or if it is zero.
func ValidateNonZeroTime(i interface{}) error {
	t, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if t.IsZero() {
		return fmt.Errorf(ErrZeroTime, t)
	}
	return nil
}

// ValidateNonZeroDuration validates if a duration is non zero
func ValidateNonZeroDuration(i interface{}) error {
	d, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if d == 0 {
		return fmt.Errorf(ErrZeroDuration, d)
	}
	return nil
}

// ValidateNonZeroDec validates if the given parameter is a non-zero
func ValidateNonZeroDec(i interface{}) error {
	d, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if d.IsZero() {
		return fmt.Errorf(ErrZeroDec, d)
	}
	return nil
}

// ValidatePositiveInt validates if the given parameter is a non-zero
func ValidatePositiveInt(i interface{}) error {
	in, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if in.IsNil() || in.LTE(math.ZeroInt()) {
		return fmt.Errorf(ErrZeroOrNegativeInt, in)
	}
	return nil
}

// ValidatePositiveCoin validates if the given parameter is a non-zero coin
func ValidatePositiveCoin(i interface{}) error {
	d, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if d.IsZero() {
		return fmt.Errorf(ErrZeroCoin, d)
	}
	// Validate the coin itself
	err := d.Validate()
	if err != nil {
		return err
	}
	return nil
}

// ValidatePositiveU32 validates if the given parameter is a non-zero uint32
func ValidatePositiveU32(i interface{}) error {
	u, ok := i.(uint32)
	if !ok {
		return fmt.Errorf(ErrInvalidType, i)
	}
	if u == 0 {
		return fmt.Errorf(ErrZeroU32, u)
	}
	return nil
}
