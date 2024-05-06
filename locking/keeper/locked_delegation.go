package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/types"
)

// GetLockedDelegation returns a specific locked delegation
func (k Keeper) GetLockedDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (lockedDelegation types.LockedDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLockedDelegationKey(delAddr, valAddr)

	bz := store.Get(key)
	if bz == nil {
		return lockedDelegation, false
	}

	k.cdc.MustUnmarshal(bz, &lockedDelegation)

	return lockedDelegation, true
}

// GetAllLockedDelegations return all the locked delegations, used for testing and genesis dump
func (k Keeper) GetAllLockedDelegations(ctx sdk.Context) (lockedDelegations []types.LockedDelegation) {
	k.IterateLockedDelegations(ctx, func(lockedDelegation types.LockedDelegation) (stop bool) {
		lockedDelegations = append(lockedDelegations, lockedDelegation)
		return false
	})
	return
}

// IterateLockedDelegations iterates through all of the locked delegations
func (k Keeper) IterateLockedDelegations(ctx sdk.Context, fn func(lockedDelegation types.LockedDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LockedDelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lockedDelegation types.LockedDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &lockedDelegation)
		if fn(lockedDelegation) {
			break
		}
	}
}

// IterateDelegatorLockedDelegations iterates through a delegator's locked delegations
func (k Keeper) IterateDelegatorLockedDelegations(ctx sdk.Context, delegator sdk.AccAddress, cb func(lockedDelegation types.LockedDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GetLockedDelegationPerDelegatorKey(delegator))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lockedDelegation types.LockedDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &lockedDelegation)
		if cb(lockedDelegation) {
			break
		}
	}
}

// SetLockedDelegation sets a locked delegation
// This doesn't add current entries to the queue
func (k Keeper) SetLockedDelegation(ctx sdk.Context, lockedDelegation types.LockedDelegation) error {
	// First we validate
	err := lockedDelegation.Validate()
	if err != nil {
		return err
	}

	// We set the delAddr and valAddr to be used to get the key
	delAddr := sdk.MustAccAddressFromBech32(lockedDelegation.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(lockedDelegation.ValidatorAddress)
	if err != nil {
		return err
	}

	// Get the store and bz
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&lockedDelegation)
	if err != nil {
		return err
	}

	key := types.GetLockedDelegationKey(delAddr, valAddr)
	store.Set(key, bz)
	return nil
}

// DeleteLockedDelegation removes a locked delegation
func (k Keeper) DeleteLockedDelegation(ctx sdk.Context, lockedDelegation types.LockedDelegation) error {
	// We set the delAddr and valAddr to be used to get the key
	delAddr := sdk.MustAccAddressFromBech32(lockedDelegation.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(lockedDelegation.ValidatorAddress)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if err != nil {
		return err
	}

	key := types.GetLockedDelegationKey(delAddr, valAddr)
	store.Delete(key)
	return nil
}

// SetLockedDelegationEntry adds an entry to the locked delegation
// It creates the locked delegation if it does not exist.
// this doesn't directly add the entry to the queue or to the look up
func (k Keeper) SetLockedDelegationEntry(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	entry types.LockedDelegationEntry,
) (types.LockedDelegation, error) {
	// Check if we already have a lockedDelegation
	lockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valAddr)
	if found {
		// If found just add
		lockedDelegation.AddEntry(entry)
	} else {
		// If not found create a new locked delegation
		lockedDelegation = types.NewLockedDelegation(
			delAddr,
			valAddr,
			[]types.LockedDelegationEntry{entry},
		)
	}

	// Set the delegation on top of the old one
	err := k.SetLockedDelegation(ctx, lockedDelegation)
	if err != nil {
		return types.LockedDelegation{}, err
	}

	return lockedDelegation, nil
}

// SetLockedDelegationByEntryID sets an index to look up an Locked Delegation by the entry ID
// it does not set the locked delegation itself, only the look up
func (k Keeper) SetLockedDelegationByEntryID(ctx sdk.Context, lockedDelegation types.LockedDelegation, id uint64) error {
	store := ctx.KVStore(k.storeKey)
	delAddr := sdk.MustAccAddressFromBech32(lockedDelegation.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(lockedDelegation.ValidatorAddress)
	if err != nil {
		return err
	}

	lockedDelegationKey := types.GetLockedDelegationKey(delAddr, valAddr)
	store.Set(types.GetLockedDelegationIndexKey(id), lockedDelegationKey)
	return nil
}

// GetLockedDelegationByEntryID returns a locked delegation that has an locked delegation entry with a certain ID
func (k Keeper) GetLockedDelegationByEntryID(ctx sdk.Context, id uint64) (lockedDelegation types.LockedDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Get the LD key from the ID
	lockedDelegationKey := store.Get(types.GetLockedDelegationIndexKey(id))
	if lockedDelegationKey == nil {
		return types.LockedDelegation{}, false
	}

	// From the LD key get the object
	bz := store.Get(lockedDelegationKey)
	if bz == nil {
		return types.LockedDelegation{}, false
	}

	// Unmarshal it, if it error out we just got the wrong object
	err := k.cdc.Unmarshal(bz, &lockedDelegation)
	if err != nil {
		return types.LockedDelegation{}, false
	}

	return lockedDelegation, true
}

// DeleteLockedDelegationIndex removes a mapping from entry id to a locked delegation
func (k Keeper) DeleteLockedDelegationIndex(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLockedDelegationIndexKey(id))
}

// IncrementLockedDelegationEntryID increments and returns a unique ID for an locked delegation entry
// We can always increment it
func (k Keeper) IncrementLockedDelegationEntryID(ctx sdk.Context) (lockedDelegationEntryID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LockedDelegationEntryIDKey)
	if bz != nil {
		lockedDelegationEntryID = binary.BigEndian.Uint64(bz)
	}

	lockedDelegationEntryID++

	// Convert back into bytes for storage
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, lockedDelegationEntryID)

	store.Set(types.LockedDelegationEntryIDKey, bz)

	return lockedDelegationEntryID
}

// SetInitialLockedDelegationEntryID set the initial next value for a locked delegation ID, used in genesis startup
func (k Keeper) SetInitialLockedDelegationEntryID(ctx sdk.Context, initialID uint64) {
	store := ctx.KVStore(k.storeKey)
	// Convert into bytes for storage
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, initialID)

	store.Set(types.LockedDelegationEntryIDKey, bz)
}
