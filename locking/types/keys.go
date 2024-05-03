// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the staking helper module
	ModuleName = "locking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the staking helper module
	RouterKey = ModuleName
)

// Parameter store keys
var (
	ParamsKey = []byte("Params")

	// Keys for store prefixes
	LockedDelegationKey = []byte{0x11} // key for a locked delegation

	// Queues
	LockedDelegationsQueueKey = []byte{0x21} // The queue for unlocking locked delegations

	// Counters
	LockedDelegationEntryIDKey = []byte{0x31} // key for the incrementing counter id for locked delegation entry id
	LockedDelegationIndexKey   = []byte{0x38} // prefix for an index for looking up locked delegation by their ID
)

// GetLockedDelegationsKey creates the prefix for a locked delegation for one validators
func GetLockedDelegationKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetLockedDelegationPerDelegatorKey(delAddr), address.MustLengthPrefix(valAddr)...)
}

// GetLockedDelegationsKey creates the prefix for all locked delegation from a delegator
func GetLockedDelegationPerDelegatorKey(delAddr sdk.AccAddress) []byte {
	return append(LockedDelegationKey, address.MustLengthPrefix(delAddr)...)
}

// GetLockedDelegationTimeKey returns a timed key for a locked delegation
// used for queuing unlocking delegations
func GetLockedDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(LockedDelegationsQueueKey, bz...)
}

// GetLockedDelegationIndexKey returns a key for the index for looking up a locked delegation by the entries it contain
func GetLockedDelegationIndexKey(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(LockedDelegationIndexKey, bz...)
}
