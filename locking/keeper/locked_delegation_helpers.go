// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

// CreateLockedDelegationEntry creates a new locked delegation
// It also adds the new entry to the queue and add it to index loop up
func (k Keeper) CreateLockedDelegationEntry(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	amount math.Int,
	rate types.Rate,
	autoRenew bool,
) (types.LockedDelegationEntry, error) {
	// Check if we have reach the limit of entries
	if k.HasMaxLockedDelegationEntries(ctx, delAddr, valAddr) {
		return types.LockedDelegationEntry{}, types.ErrMaxLockedDelegationEntriesReached
	}

	// Create a new entry with a new ID
	unlockOn := ctx.BlockTime().Add(rate.Duration)
	id := k.IncrementLockedDelegationEntryID(ctx)

	// Fetch the validator and calculate the new shares
	// If not found we can use a empty validator for calculation
	validator, _ := k.stakingKeeper.GetValidator(ctx, valAddr)

	// calculate the shares
	shares := types.CalculateSharesFromValidator(amount, validator)

	entry := types.NewLockedDelegationEntry(
		shares,
		rate,
		unlockOn,
		autoRenew,
		id,
	)
	if err := entry.Validate(); err != nil {
		return types.LockedDelegationEntry{}, err
	}

	// Store the entry on the system
	lockedDelegation, err := k.SetLockedDelegationEntry(ctx, delAddr, valAddr, entry)
	if err != nil {
		return types.LockedDelegationEntry{}, err
	}

	// Add it to the queue
	k.InsertLockedDelegationQueue(ctx, lockedDelegation, entry.UnlockOn)

	// Add it to the index look up
	err = k.SetLockedDelegationByEntryID(ctx, lockedDelegation, id)
	if err != nil {
		return types.LockedDelegationEntry{}, err
	}

	return entry, nil
}

// CreateLockedDelegationEntryAndDelegate creates a new locked delegation entry and a new delegation on top
func (k Keeper) CreateLockedDelegationEntryAndDelegate(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	amount math.Int,
	lockDuration time.Duration,
	autoRenew bool,
) (types.LockedDelegationEntry, error) {
	// Check if the selected rate exists
	params := k.GetParams(ctx)
	rate, found := params.GetRateFromDuration(lockDuration)
	if !found {
		return types.LockedDelegationEntry{}, types.ErrCreateLockedDelegationDurationUnmatch
	}

	// Check the validator for the delegation
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.LockedDelegationEntry{}, stakingtypes.ErrNoValidatorFound
	}

	// We first create the locked delegation
	// This must be done before the delegation to use the shares before it's updated
	entry, err := k.CreateLockedDelegationEntry(
		ctx,
		delAddr,
		valAddr,
		amount,
		rate,
		autoRenew,
	)
	if err != nil {
		return types.LockedDelegationEntry{}, err
	}

	// Now we create the normal delegation using the staking delegate
	// This checks the validator status and does the token management
	// If this passes with no issues and a delegation is created,
	// we can safely create the lock
	_, err = k.stakingKeeper.Delegate(
		ctx,
		delAddr,
		amount,
		stakingtypes.Unbonded,
		validator,
		true,
	)
	if err != nil {
		return types.LockedDelegationEntry{}, err
	}

	return entry, nil
}

// LockedDelegationRedelegation move the locked delegations from one validator to another
// This also sets the new queue for the target validator
func (k Keeper) LockedDelegationRedelegation(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valSrcAddr,
	valDstAddr sdk.ValAddress,
	ids []uint64,
) (math.LegacyDec, math.Int, error) {
	// Prepare the locked delegation and get the used values
	// This also does any necessary rewards collection
	srcLockedDelegation, srcValidator, dstValidator, err := k.prepareLockedDelegationRedelegation(
		ctx,
		delAddr,
		valSrcAddr,
		valDstAddr,
	)
	if err != nil {
		return math.LegacyDec{}, math.Int{}, err
	}

	// Check if all the requested IDs can be found on the src locked delegation
	exists, foundSrcEntries := srcLockedDelegation.EntriesForIds(ids)
	if !exists {
		return math.LegacyDec{}, math.Int{}, types.ErrLockedDelegationEntryNotFound
	}

	// Now apply the real redelegate
	var dstLockedDelegation types.LockedDelegation
	tokensMoved := math.ZeroInt()
	sharesMoved := math.LegacyZeroDec()
	for _, entry := range foundSrcEntries {
		// Delete the old ID from lookup, since we are changing validators
		k.DeleteLockedDelegationIndex(ctx, entry.Id)

		// Simulate the share removal
		tokensToRemove := types.SimulateValidatorSharesRemoval(entry.Shares, srcValidator)

		// Update the tokens and shares removed
		tokensMoved = tokensMoved.Add(tokensToRemove)
		sharesMoved = sharesMoved.Add(entry.Shares)

		// Update the share value
		// This is necessary, since the new validator may use a different share ratio
		shares := types.CalculateSharesFromValidator(tokensToRemove, dstValidator)
		if err != nil {
			return math.LegacyDec{}, math.Int{}, err
		}
		entry.Shares = shares

		// Store the new locked delegation entries on top of the destination validator
		dstLockedDelegation, err = k.SetLockedDelegationEntry(ctx, delAddr, valDstAddr, entry)
		if err != nil {
			return math.LegacyDec{}, math.Int{}, err
		}
		// Add it to the queue
		k.InsertLockedDelegationQueue(ctx, dstLockedDelegation, entry.UnlockOn)
		// Add it to the look up
		err = k.SetLockedDelegationByEntryID(ctx, dstLockedDelegation, entry.Id)
		if err != nil {
			return math.LegacyDec{}, math.Int{}, err
		}
	}

	// Delete all the ids from the src locked delegation
	srcLockedDelegation.RemoveEntries(foundSrcEntries)

	// Delete the empty locked delegation if empty
	// If not, update the old entry
	if len(srcLockedDelegation.Entries) == 0 {
		err = k.DeleteLockedDelegation(ctx, srcLockedDelegation)
	} else {
		err = k.SetLockedDelegation(ctx, srcLockedDelegation)
	}
	if err != nil {
		return math.LegacyDec{}, math.Int{}, err
	}

	return sharesMoved, tokensMoved, nil
}

// prepareLockedDelegationRedelegation prepares locked delegations for redelegation
// it does all the necessary validations and the extract of current rewards
func (k Keeper) prepareLockedDelegationRedelegation(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valSrcAddr,
	valDstAddr sdk.ValAddress,
) (srcLockedDelegation types.LockedDelegation, srcValidator, dstValidator stakingtypes.Validator, err error) {
	// Check if we will reach the max entries
	if k.LDRedelegationWillReachMaxEntries(ctx, delAddr, valSrcAddr, valDstAddr) {
		return types.LockedDelegation{}, stakingtypes.Validator{}, stakingtypes.Validator{}, types.ErrLDRedelegationMaxEntriesReached
	}

	// Get the source locked delegation and entries
	srcLockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valSrcAddr)
	if !found {
		return types.LockedDelegation{}, stakingtypes.Validator{}, stakingtypes.Validator{}, types.ErrLockedDelegationNotFound
	}

	// Get both validators
	// It will be used to calculate the new shares
	srcValidator, _ = k.stakingKeeper.GetValidator(ctx, valSrcAddr)
	dstValidator, _ = k.stakingKeeper.GetValidator(ctx, valDstAddr)

	// Check the shares that will be updated
	if srcLockedDelegation.TotalShares().IsZero() || srcValidator.GetTokens().IsZero() {
		return types.LockedDelegation{}, stakingtypes.Validator{}, stakingtypes.Validator{}, types.ErrLockedDelegationRedelegationZeroShares
	}

	// Do a rewards withdraw before anything get's updated
	_, err = k.distributionKeeper.WithdrawDelegationRewards(ctx, delAddr, valSrcAddr)
	if err != nil {
		return types.LockedDelegation{}, stakingtypes.Validator{}, stakingtypes.Validator{}, err
	}
	// Only apply to dst validator if we have a delegation
	_, dstDelegationFound := k.stakingKeeper.GetDelegation(ctx, delAddr, valDstAddr)
	if dstDelegationFound {
		_, err = k.distributionKeeper.WithdrawDelegationRewards(ctx, delAddr, valDstAddr)
		if err != nil {
			return types.LockedDelegation{}, stakingtypes.Validator{}, stakingtypes.Validator{}, err
		}
	}

	return srcLockedDelegation, srcValidator, dstValidator, nil
}

// LockedDelegationAndStakingRedelegation creates a locked delegation redelegation and a staking redelegation
func (k Keeper) LockedDelegationAndStakingRedelegation(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valSrcAddr,
	valDstAddr sdk.ValAddress,
	ids []uint64,
) (math.LegacyDec, math.Int, error) {
	// Get the total locked shares for the pair
	// If the shares are zero, we don't have anything to redelegate
	lockedShares := k.LockedDelegationTotalShares(
		ctx,
		delAddr,
		valSrcAddr,
	)
	if lockedShares.IsZero() {
		return math.LegacyDec{}, math.Int{}, types.ErrNoLockedSharesToRedelegate
	}

	// Now we can do the locked delegation redelegation
	sharesMoved, tokensMoved, err := k.LockedDelegationRedelegation(
		ctx,
		delAddr,
		valSrcAddr,
		valDstAddr,
		ids,
	)
	if err != nil {
		return math.LegacyDec{}, math.Int{}, err
	}

	// Do the normal staking redelegate operation
	_, err = k.stakingKeeper.BeginRedelegation(
		ctx, delAddr, valSrcAddr, valDstAddr, sharesMoved,
	)
	if err != nil {
		return math.LegacyDec{}, math.Int{}, err
	}

	return sharesMoved, tokensMoved, nil
}

// ToggleLockedDelegationEntryAutoRenew toggles a locked delegation entry based on the entry Id
func (k Keeper) ToggleLockedDelegationEntryAutoRenew(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	entryID uint64,
) (entry types.LockedDelegationEntry, err error) {
	// Find the locked delegation
	lockedDelegation, found := k.GetLockedDelegation(
		ctx, delAddr, valAddr,
	)
	if !found {
		return entry, types.ErrLockedDelegationNotFound
	}

	// Switch the auto renew field
	entry, found = lockedDelegation.ToggleAutoRenewForID(entryID)
	if !found {
		return entry, types.ErrLockedDelegationEntryNotFound
	}

	// Save the locked delegation
	// We don't need to update the queue, since the unlock on has not changed
	// And since the ID haven't changed, we don't need to update the look up
	err = k.SetLockedDelegation(ctx, lockedDelegation)
	if err != nil {
		return entry, err
	}

	return entry, nil
}

// HasMaxUnbondingDelegationEntries check if unbonding delegation has maximum number of entries.
func (k Keeper) HasMaxLockedDelegationEntries(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) bool {
	lockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return false
	}

	return len(lockedDelegation.Entries) >= int(k.MaxEntries(ctx))
}

// LDRedelegationWillReachMaxEntries check if a redelegation will reach the max entries for the target locked delegation
func (k Keeper) LDRedelegationWillReachMaxEntries(ctx sdk.Context, delAddr sdk.AccAddress, srcValAddr sdk.ValAddress, dstValAddr sdk.ValAddress) bool {
	srcLockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, srcValAddr)
	if !found {
		return false
	}
	dstLockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, dstValAddr)
	if !found {
		return false
	}

	return len(srcLockedDelegation.Entries)+len(dstLockedDelegation.Entries) >= int(k.MaxEntries(ctx))
}

// LockedDelegationTotalShares returns the total shares a delegator validator pair has
func (k Keeper) LockedDelegationTotalShares(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) math.LegacyDec {
	lockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return math.LegacyZeroDec()
	}

	return lockedDelegation.TotalShares()
}

// LockedDelegationTotalTokens returns the total tokens a delegator validator pair has
func (k Keeper) LockedDelegationTotalTokens(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) math.LegacyDec {
	total := k.LockedDelegationTotalShares(ctx, delAddr, valAddr)

	// Calculate the tokens based on the validator shares
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		validator = stakingtypes.Validator{}
	}
	shares := validator.TokensFromShares(total)

	return shares
}

// GetDelegatorLockedShares returns the total shares a delegator has locked
func (k Keeper) GetDelegatorLockedShares(ctx sdk.Context, delegator sdk.AccAddress) math.LegacyDec {
	lockedShares := math.LegacyZeroDec()
	k.IterateDelegatorLockedDelegations(ctx, delegator, func(lockedDelegation types.LockedDelegation) bool {
		lockedShares = lockedDelegation.TotalShares()
		return false
	})
	return lockedShares
}

// DequeueExpiredLockedDelegations dequeue and unlock delegations
func (k Keeper) DequeueExpiredLockedDelegations(ctx sdk.Context, currTime time.Time) (expiredPairs []types.LockedDelegationPair) {
	store := ctx.KVStore(k.storeKey)

	// gets an iterator for all timeslices from time 0 until the current time
	iterator := k.LockedDelegationQueueIterator(ctx, currTime)
	defer iterator.Close()

	// Get all expired pairs and dequeue
	for ; iterator.Valid(); iterator.Next() {
		// Unmarshal the values
		pairsAtTime := types.LockedDelegationPairs{}
		value := iterator.Value()
		k.cdc.MustUnmarshal(value, &pairsAtTime)

		// Check uniques for a single iteration
		expiredPairs = append(expiredPairs, pairsAtTime.UniquePairs()...)

		// Delete this old key
		store.Delete(iterator.Key())
	}

	return
}

// CompleteLockedDelegations processes locked delegations and unlocks or renews them
// based on their configuration when their time has expired
func (k Keeper) CompleteLockedDelegations(ctx sdk.Context, pair types.LockedDelegationPair) error {
	delAddr := sdk.MustAccAddressFromBech32(pair.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(pair.ValidatorAddress)
	if err != nil {
		return err
	}

	// Get the locked delegation
	lockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil
	}

	// Process the entries
	// Remove expired entries and renew the ones needed
	totalUndelegate, err := k.processEntries(ctx, &lockedDelegation)
	if err != nil {
		return err
	}

	// Before updating the delegation, we must collect the rewards using the current locked delegation in store
	// This avoids losing the total locked delegation reward when undelegating
	// We also only need to apply if the total to undelegation isn't zero
	if !totalUndelegate.IsZero() {
		_, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
		if err != nil {
			return err
		}

		// Before updating the locked delegation, let's check if we can undelegate
		delegation := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
		if delegation.GetShares().LT(totalUndelegate) {
			return types.ErrLockedSharesSmallerThanDelegation
		}
	}

	// Update or delete the locked delegation depending on its entries
	// set the redelegation or remove it if there are no more entries
	if len(lockedDelegation.Entries) == 0 {
		err := k.DeleteLockedDelegation(ctx, lockedDelegation)
		if err != nil {
			return err
		}
	} else {
		err := k.SetLockedDelegation(ctx, lockedDelegation)
		if err != nil {
			return err
		}
	}

	// Undelegate the expected amount
	// We want to undelegate as the last action to avoid conflicts with the hooks
	if !totalUndelegate.IsZero() {
		_, err := k.stakingKeeper.Undelegate(ctx, delAddr, valAddr, totalUndelegate)
		if err != nil {
			return err
		}
	}

	return nil
}

// processEntries processes locked delegation entries by renewing or removing them
// It returns the total amount to be undelegated
func (k Keeper) processEntries(ctx sdk.Context, ld *types.LockedDelegation) (math.LegacyDec, error) {
	currTime := ctx.BlockTime()
	totalUndelegate := math.LegacyZeroDec()

	// Iterate over the entries
	// here we must use indexing due to the list removal or addition
	// New items are added last, and we shouldn't check them
	for i := 0; i < len(ld.Entries); i++ {
		entry := ld.Entries[i]

		if !entry.Expired(currTime) {
			continue
		}

		// Remove the expired entry
		ld.RemoveEntryForIndex(int64(i))
		i--

		// Remove the ID from look up
		k.DeleteLockedDelegationIndex(ctx, entry.Id)

		// Check if we should renew
		if entry.AutoRenew {
			// Handle auto-renew process
			err := k.handleAutoRenew(ctx, ld, entry)
			if err != nil {
				return math.LegacyZeroDec(), err
			}
		} else {
			// Add the undelegate total if we don't auto renew
			totalUndelegate = totalUndelegate.Add(entry.Shares)
		}
	}

	return totalUndelegate, nil
}

// handleAutoRenew handles the auto-renewal process for a given entry
func (k Keeper) handleAutoRenew(ctx sdk.Context, ld *types.LockedDelegation, entry types.LockedDelegationEntry) error {
	entry.UnlockOn = entry.UnlockOn.Add(entry.Rate.Duration)
	// Add the entry
	ld.AddEntry(entry)
	// Add to the queue
	k.InsertLockedDelegationQueue(ctx, *ld, entry.UnlockOn)
	// Add to look up
	return k.SetLockedDelegationByEntryID(ctx, *ld, entry.Id)
}
