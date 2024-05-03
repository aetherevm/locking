// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// TODO: Remove the rewards denom, we should use the staking bonding denom

const (
	ErrModuleEnabledInvalid = "%s enabled is invalid: %s"
	ErrDenomInvalid         = "%s denom is invalid: %s"
	ErrMaxEntriesInvalid    = "%s max entries is invalid: %s"

	// Rate errors
	ErrRateDurationInvalid = "%s rate duration is invalid: %s"
	ErrRateDecInvalid      = "%s rate dec is invalid: %s"
	ErrRateNotUnique       = "%s rate duration of %s not unique for the current rates"
)

var (
	// DefaultMaxEntries is the max of locked delegations a val/del can have
	DefaultMaxEntries uint32 = 100

	// DefaultRates defines the default rates of the reward system
	DefaultRates = []Rate{
		// 8 months lock 2.2% reward
		NewRate(8*30*24*time.Hour, sdk.NewDecWithPrec(22, 1)),
		// 16 months lock 3.3% reward
		NewRate(16*30*24*time.Hour, sdk.NewDecWithPrec(33, 1)),
		// 24 months lock 4.4% reward
		NewRate(24*30*24*time.Hour, sdk.NewDecWithPrec(44, 1)),
		// 32 months lock 5.5% reward
		NewRate(32*30*24*time.Hour, sdk.NewDecWithPrec(55, 1)),
	}
)

// NewParams returns a new param
func NewParams(
	maxEntries uint32, rates []Rate,
) Params {
	return Params{
		MaxEntries: maxEntries,
		Rates:      rates,
	}
}

// DefaultParams returns the default params
func DefaultParams() Params {
	return Params{
		MaxEntries: DefaultMaxEntries,
		Rates:      DefaultRates,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := ValidatePositiveU32(p.MaxEntries); err != nil {
		return fmt.Errorf(ErrMaxEntriesInvalid, ModuleName, err)
	}

	// Validate the rates, rate duration also should be unique
	seenDurations := make(map[time.Duration]bool)
	for _, rate := range p.Rates {
		if err := rate.Validate(); err != nil {
			return err
		}

		// Check for uniqueness on the duration
		if _, exists := seenDurations[rate.Duration]; exists {
			return fmt.Errorf(ErrRateNotUnique, ModuleName, rate.Duration)
		}
		seenDurations[rate.Duration] = true
	}
	return nil
}

// String returns the string representation of Params
func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		return ""
	}
	return string(out)
}

// GetRateFromDuration returns a rate based on a input duration
func (p Params) GetRateFromDuration(duration time.Duration) (rate Rate, found bool) {
	for _, rate := range p.Rates {
		if rate.Duration == duration {
			return rate, true
		}
	}
	return Rate{}, false
}

// NewRate returns a new rate
func NewRate(
	duration time.Duration, rate sdk.Dec,
) Rate {
	return Rate{
		Duration: duration,
		Rate:     rate,
	}
}

// Validate validates a single rate
func (r Rate) Validate() error {
	if err := ValidateNonZeroDuration(r.Duration); err != nil {
		return fmt.Errorf(ErrRateDurationInvalid, ModuleName, err)
	}
	if err := ValidateNonZeroDec(r.Rate); err != nil {
		return fmt.Errorf(ErrRateDecInvalid, ModuleName, err)
	}
	return nil
}
