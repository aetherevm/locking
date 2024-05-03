// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types

import (
	fmt "fmt"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"gopkg.in/yaml.v2"
)

const (
	ErrDelegatorAddressInvalid = "%s invalid delegator address: %s"
	ErrValidatorAddressInvalid = "%s invalid validator address: %s"
	ErrRateInvalid             = "%s invalid rate: %s"
	ErrSharesInvalid           = "%s invalid shares: %s"
	ErrLockDurationInvalid     = "%s invalid lock duration: %s"
	ErrUnlockOnInvalid         = "%s invalid unlock time: %s"
	ErrAuthorityInvalid        = "%s invalid authority address: %s"

	ErrEntryNotUnique = "%s locked delegation entry not unique: %s"
)

// NewLockedDelegation return a new LockedDelegation
func NewLockedDelegation(
	delegatorAddress sdk.AccAddress,
	validatorAddress sdk.ValAddress,
	entries []LockedDelegationEntry,
) LockedDelegation {
	return LockedDelegation{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Entries:          entries,
	}
}

// Validate validates a LockedDelegation
func (ld LockedDelegation) Validate() error {
	if _, err := sdk.AccAddressFromBech32(ld.DelegatorAddress); err != nil {
		return fmt.Errorf(ErrDelegatorAddressInvalid, ModuleName, err)
	}
	if _, err := sdk.ValAddressFromBech32(ld.ValidatorAddress); err != nil {
		return fmt.Errorf(ErrValidatorAddressInvalid, ModuleName, err)
	}

	// Validate all the entries one at a time and check for duplicates
	seenLockedDelegationIDs := make(map[uint64]bool)
	for _, lockedDelegationEntry := range ld.Entries {
		if err := lockedDelegationEntry.Validate(); err != nil {
			return err
		}

		if _, exists := seenLockedDelegationIDs[lockedDelegationEntry.Id]; exists {
			return fmt.Errorf(ErrEntryNotUnique, ModuleName, lockedDelegationEntry)
		}
		seenLockedDelegationIDs[lockedDelegationEntry.Id] = true
	}
	return nil
}

// String returns a human readable string representation of a Delegation.
func (ld LockedDelegation) String() string {
	out, _ := yaml.Marshal(ld)
	return string(out)
}

// GetDelegatorAddr returns the delegator address as sdk.AccAddress
func (ld LockedDelegation) GetDelegatorAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(ld.DelegatorAddress)
}

// GetValidatorAddr returns the validator address as sdk.ValAddress
func (ld LockedDelegation) GetValidatorAddr() (sdk.ValAddress, error) {
	addr, err := sdk.ValAddressFromBech32(ld.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// AddEntry - append entry to the locked delegation
func (ld *LockedDelegation) AddEntry(entry LockedDelegationEntry) {
	index := -1

	// Let's say that we have one entry with the same values
	// First find the index
	for i, currentEntry := range ld.Entries {
		if currentEntry.Rate.Equal(&entry.Rate) &&
			currentEntry.AutoRenew == entry.AutoRenew &&
			currentEntry.UnlockOn == entry.UnlockOn {
			index = i
		}
	}

	// Now based on the index we update the values of the old entry or just add it
	if index != -1 {
		ldEntry := ld.Entries[index]
		ldEntry.Shares = ldEntry.Shares.Add(entry.Shares)

		// Update the entry
		ld.Entries[index] = ldEntry
	} else {
		// If we don't find we just append
		ld.Entries = append(ld.Entries, entry)
	}
}

// RemoveEntryForIndex removes a single locked delegation entry from the entries list based on the index
func (ld *LockedDelegation) RemoveEntryForIndex(index int64) {
	if index < 0 || index > int64(len(ld.Entries)-1) {
		return
	}
	ld.Entries = append(ld.Entries[:index], ld.Entries[index+1:]...)
}

// RemoveEntries removes a locked delegation entries from the entries list based on ids
func (ld *LockedDelegation) RemoveEntries(entries []LockedDelegationEntry) {
	idSet := make(map[uint64]bool)
	for _, entry := range entries {
		idSet[entry.Id] = true
	}

	// Create a new list for entries that are not removed
	newEntries := make([]LockedDelegationEntry, 0)
	for _, entry := range ld.Entries {
		_, exists := idSet[entry.Id]
		// Add to the list if the id doesn't exists in the expected removal
		if !exists {
			newEntries = append(newEntries, entry)
		}
	}

	// Update the Entries field
	ld.Entries = newEntries
}

// TotalShares is the total of shares on top of that locked delegation
func (ld LockedDelegation) TotalShares() math.LegacyDec {
	shares := math.LegacyZeroDec()
	for _, entry := range ld.Entries {
		shares = shares.Add(entry.Shares)
	}
	return shares
}

// WeightedRatio calculates the ratio for all locked delegation entries
// the value is calculated by the sum of locked shares multiplied by the ratio, divided by the total shares locked
// for this function was considering units until e-20
func (ld LockedDelegation) WeightedRatio() (weightedRatio math.LegacyDec) {
	lockedSum := math.LegacyZeroDec()
	weight := math.LegacyZeroDec()

	for _, entry := range ld.Entries {
		// Improve the precision on weight calculation
		rate := entry.Rate.Rate.MulInt64(10)

		weight = weight.Add(entry.Shares.Mul(rate))
		lockedSum = lockedSum.Add(entry.Shares)
	}

	// We don't want to divide by zero
	if lockedSum.IsZero() {
		return math.LegacyZeroDec()
	}

	return weight.Quo(lockedSum).QuoInt64(1000)
}

// CalculateDelegationRatio calculates the ratio of a locked delegation based on a delegation share
// This calculates a ratio between the delegation shares and the sum of locks and multiply by the weightedRatio
func (ld LockedDelegation) CalculateDelegationRatio(shares math.LegacyDec) (ratio math.LegacyDec) {
	weightedRatio := ld.WeightedRatio()

	totalShares := ld.TotalShares()
	// We don't want to divide by zero
	if totalShares.IsZero() || shares.IsZero() {
		return math.LegacyZeroDec()
	}

	delegationRatio := totalShares.Quo(shares)

	// lock the max ratio between delegation shares and locked delegation shares in one
	if delegationRatio.GT(math.LegacyOneDec()) {
		delegationRatio = math.LegacyOneDec()
	}

	return delegationRatio.Mul(weightedRatio)
}

// ToggleAutoRenewForID - toggle a entry auto renew based on it's id
func (ld *LockedDelegation) ToggleAutoRenewForID(id uint64) (entry LockedDelegationEntry, found bool) {
	// Find the entry and update the locked delegation auto renew
	for i, currentEntry := range ld.Entries {
		if currentEntry.Id == id {
			ld.Entries[i].AutoRenew = !ld.Entries[i].AutoRenew
			return ld.Entries[i], true
		}
	}

	return LockedDelegationEntry{}, false
}

// EntriesForIds checks if the list of Ids exists as locked delegation entries
// returns false if a the locked delegation if empty of a single entry doesn't exists
// returns the entries if all ids exists
func (ld LockedDelegation) EntriesForIds(ids []uint64) (exists bool, entries []LockedDelegationEntry) {
	// If the locked delegation is empty we can just return false
	if len(ld.Entries) == 0 {
		return false, nil
	}

	// If the list of ids is empty, we can return all the available entries
	if len(ids) == 0 {
		return true, ld.Entries
	}

	// Iterate over the entries filling the entriesPerIDMap
	entriesPerIDMap := make(map[uint64]LockedDelegationEntry)
	for _, entry := range ld.Entries {
		entriesPerIDMap[entry.Id] = entry
	}

	// Now iterate over the requested list to find non existing IDs
	// And build the return list
	entriesFound := make([]LockedDelegationEntry, 0, len(ids))
	computedIds := make(map[uint64]bool)
	for _, id := range ids {
		// Check if this ID has been checked
		_, found := computedIds[id]
		if found {
			continue
		}

		// Try to find the entry, if found add to our found entries list
		entryFound, found := entriesPerIDMap[id]
		if !found {
			return false, nil
		}
		entriesFound = append(entriesFound, entryFound)

		// Add to our computed IDs map
		computedIds[id] = true
	}

	return true, entriesFound
}

// NewLockedDelegationEntry returns a new locked delegation entry
func NewLockedDelegationEntry(
	shares math.LegacyDec,
	rate Rate,
	unlockOn time.Time,
	autoRenew bool,
	id uint64,
) LockedDelegationEntry {
	return LockedDelegationEntry{
		Shares:    shares,
		Rate:      rate,
		UnlockOn:  unlockOn,
		AutoRenew: autoRenew,
		Id:        id,
	}
}

// Validate validates a LockedDelegationEntry
func (lde LockedDelegationEntry) Validate() error {
	if err := ValidateNonZeroDec(lde.Shares); err != nil {
		return fmt.Errorf(ErrSharesInvalid, ModuleName, err)
	}
	if err := lde.Rate.Validate(); err != nil {
		return fmt.Errorf(ErrRateInvalid, ModuleName, err)
	}
	if err := ValidateNonZeroTime(lde.UnlockOn); err != nil {
		return fmt.Errorf(ErrUnlockOnInvalid, ModuleName, err)
	}
	return nil
}

// String returns a human readable string representation of a Delegation.
func (lde LockedDelegationEntry) String() string {
	out, _ := yaml.Marshal(lde)
	return string(out)
}

func (lde LockedDelegationEntry) Expired(currentTime time.Time) bool {
	return !currentTime.Before(lde.UnlockOn)
}

// String implements the Stringer interface for a LockedDelegationPair object
func (dv LockedDelegationPair) String() string {
	out, _ := yaml.Marshal(dv)
	return string(out)
}

// String implements the Stringer interface for a LockedDelegationPair object
func (dv LockedDelegationPairs) UniquePairs() []LockedDelegationPair {
	uniquePairsMap := make(map[string]LockedDelegationPair)
	for _, pair := range dv.Pairs {
		// Save unique pairs
		key := pair.String()
		uniquePairsMap[key] = pair
	}

	// Convert map back to slice
	uniquePairs := make([]LockedDelegationPair, 0, len(dv.Pairs))
	for _, pair := range uniquePairsMap {
		uniquePairs = append(uniquePairs, pair)
	}

	return uniquePairs
}

// CalculateSharesForLDEntry calculates the shares for a locked delegation entry based on a validator
// this function is based on AddTokensFromDel from the validator interface
func CalculateSharesFromValidator(amount math.Int, validator stakingtypes.Validator) math.LegacyDec {
	if validator.DelegatorShares.IsNil() || validator.DelegatorShares.IsZero() {
		// the first delegation the initial amount is the shares
		return math.LegacyNewDecFromInt(amount)
	}
	// The error here is insufficient shares
	// We can throw it away
	shares, err := validator.SharesFromTokens(amount)
	if err != nil {
		return math.LegacyZeroDec()
	}

	return shares
}

// SimulateValidatorSharesRemoval simulates a share removal, it's used on redelegating from one validator to another
// this function is based on removeTokens from the validator interface
func SimulateValidatorSharesRemoval(delShares sdk.Dec, validator stakingtypes.Validator) math.Int {
	if validator.DelegatorShares.IsNil() || validator.DelegatorShares.IsZero() {
		// the first delegation the initial amount is the shares
		return math.ZeroInt()
	}

	remainingShares := validator.DelegatorShares.Sub(delShares)

	var removedTokens math.Int
	if remainingShares.IsZero() {
		removedTokens = validator.Tokens
	} else {
		removedTokens = validator.TokensFromShares(delShares).TruncateInt()
	}

	return removedTokens
}
