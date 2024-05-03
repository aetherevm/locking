// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// TODO: BE CAREFUL WITH THE HOOKS! WE DON'T WANT TO BREAK CHAIN STATE

// For handling locking, we have to handle moments on the delegation life-cycle:
// when a delegation is modified it never should go bellow the total locked shares.
// To fully handle this situation, two paths are shown by the staking undelegate tx:
// 1. AfterDelegationModified:
// - Triggered when the delegation is partially undelegated and the final shares are non-zero.
// - This hook captures the updated value of the delegation after the modification.
// - We validate here to ensure that the updated delegation (post-modification) is greater
//   than or equal to the total amount of locked shares. This check is vital for maintaining
//   the integrity of the locking mechanism.
// 2. BeforeDelegationRemoved:
// - Triggered when the delegation is fully undelegated, resulting in zero final shares.
// - This hook captures the state of the delegation just before it's fully removed.
// - In cases where there's a lock on the delegation, this step is crucial to prevent
//   the complete removal of the delegation. Essentially, we can block the removal process
//   here if there are still locked shares associated with the delegation.

// Wrapper struct
type StakingHooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = StakingHooks{}

// Create new distribution hooks
func (k Keeper) StakingHooks() StakingHooks {
	return StakingHooks{k}
}

// AfterDelegationModified implements types.StakingHooks
func (h StakingHooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// Sanity check, we must enforce that the delegation never
	// goes bellow the delegated amount

	// Get the locked delegation and the total shares
	lockedDelegation, found := h.k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil
	}
	lockedShares := lockedDelegation.TotalShares()

	// Get the delegation
	delegation, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil
	}

	// Check if it's smaller than the locked shares
	if delegation.Shares.LT(lockedShares) {
		return types.ErrLockedSharesSmallerThanDelegation
	}

	return nil
}

// BeforeDelegationRemoved implements types.StakingHooks
func (h StakingHooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// We can block the delegation removal if we have a lock

	// Get the locked delegation and the total shares
	lockedDelegation, found := h.k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil
	}
	lockedShares := lockedDelegation.TotalShares()

	// Block the delegation removal if a lock exists
	if !lockedShares.IsZero() {
		return types.ErrLockedSharesSmallerThanDelegation
	}

	return nil
}

// BeforeDelegationCreated implements types.StakingHooks
func (StakingHooks) BeforeDelegationCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	// TODO implement this hook if moved to a provider chain
	// Do we really need to implement this?
	return nil
}

// BeforeDelegationSharesModified implements types.StakingHooks
func (StakingHooks) BeforeDelegationSharesModified(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	// TODO implement this hook if moved to a provider chain
	// Before modifying delegations we need to check if we never reach the locking value
	// Check the behavior first
	return nil
}

// AfterUnbondingInitiated implements types.StakingHooks
func (StakingHooks) AfterUnbondingInitiated(_ sdk.Context, _ uint64) error {
	return nil
}

// AfterValidatorBeginUnbonding implements types.StakingHooks
func (StakingHooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements types.StakingHooks
func (StakingHooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements types.StakingHooks
func (StakingHooks) AfterValidatorCreated(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved implements types.StakingHooks
func (StakingHooks) AfterValidatorRemoved(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeValidatorModified implements types.StakingHooks
func (StakingHooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed implements types.StakingHooks
func (StakingHooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ math.LegacyDec) error {
	return nil
}
