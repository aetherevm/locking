// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/types"
)

// GetLockedDelegationQueueTimeSlice returns the locked pairs at a set timestamp
func (k Keeper) GetLockedDelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (lockedDelegationPairs []types.LockedDelegationPair) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetLockedDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.LockedDelegationPair{}
	}

	// We need the pairs struct to Unmarshal
	pairs := types.LockedDelegationPairs{}
	k.cdc.MustUnmarshal(bz, &pairs)

	return pairs.Pairs
}

// SetLockedDelegationQueueTimeSlice sets the locked pairs at a set timestamp
func (k Keeper) SetLockedDelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time, lockedDelegationPairs []types.LockedDelegationPair) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.LockedDelegationPairs{Pairs: lockedDelegationPairs})
	store.Set(types.GetLockedDelegationTimeKey(timestamp), bz)
}

// InsertLockedDelegationQueue insert a locked delegation in the queue
func (k Keeper) InsertLockedDelegationQueue(ctx sdk.Context, lockedDelegation types.LockedDelegation, unlockOn time.Time) {
	// Build a pair
	lockedDelegationPair := types.LockedDelegationPair{
		DelegatorAddress: lockedDelegation.DelegatorAddress,
		ValidatorAddress: lockedDelegation.ValidatorAddress,
	}

	// Get all the pairs at that time
	pairsAtTime := k.GetLockedDelegationQueueTimeSlice(ctx, unlockOn)
	if len(pairsAtTime) == 0 {
		// If we don't have pairs create a new pair
		k.SetLockedDelegationQueueTimeSlice(ctx, unlockOn, []types.LockedDelegationPair{lockedDelegationPair})
	} else {
		// If not we can append to the current pairs
		pairsAtTime = append(pairsAtTime, lockedDelegationPair)
		k.SetLockedDelegationQueueTimeSlice(ctx, unlockOn, pairsAtTime)
	}
}

// LockedDelegationQueueIterator returns a iterator for locked delegation queue pairs from time 0 until endTime.
func (k Keeper) LockedDelegationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.LockedDelegationsQueueKey,
		sdk.InclusiveEndBytes(types.GetLockedDelegationTimeKey(endTime)))
}

// GetAllLockedDelegationQueuePairs returns all the locked delegation queue pairs from time 0 until endTime.
func (k Keeper) GetAllLockedDelegationQueuePairs(ctx sdk.Context, endTime time.Time) (pairs []types.LockedDelegationPair) {
	iterator := k.LockedDelegationQueueIterator(ctx, endTime)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		pairsAtTime := types.LockedDelegationPairs{}
		value := iterator.Value()
		k.cdc.MustUnmarshal(value, &pairsAtTime)

		// Save
		pairs = append(pairs, pairsAtTime.Pairs...)
	}
	return
}
