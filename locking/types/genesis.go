// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package types

import (
	"encoding/json"
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	ErrLDNotUnique        = "%s locked delegation not unique: %s"
	ErrLDEntryIDNotUnique = "%s locked delegation entry ID not unique: %d"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	params Params,
	lockedDelegations []LockedDelegation,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		LockedDelegations: lockedDelegations,
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation
// Validates if each locked delegation is unique
func (gs GenesisState) Validate() error {
	// We should not have duplicated LD or IDS
	seeingLD := make(map[string]map[string]bool)
	seeingLDEntryID := make(map[uint64]bool)
	for _, lockedDelegation := range gs.LockedDelegations {
		if err := lockedDelegation.Validate(); err != nil {
			return err
		}

		// Check if the inner map exists, if not, initialize it
		if _, exists := seeingLD[lockedDelegation.DelegatorAddress]; !exists {
			seeingLD[lockedDelegation.DelegatorAddress] = make(map[string]bool)
		}

		if _, exists := seeingLD[lockedDelegation.DelegatorAddress][lockedDelegation.ValidatorAddress]; exists {
			return fmt.Errorf(ErrLDNotUnique, ModuleName, lockedDelegation.String())
		}
		seeingLD[lockedDelegation.DelegatorAddress][lockedDelegation.ValidatorAddress] = true

		// Check for duplicated ID
		for _, entry := range lockedDelegation.Entries {
			if _, exists := seeingLDEntryID[entry.Id]; exists {
				return fmt.Errorf(ErrLDEntryIDNotUnique, ModuleName, entry.Id)
			}
			seeingLDEntryID[entry.Id] = true
		}
	}
	return gs.Params.Validate()
}

// GetGenesisStateFromAppState return GenesisState
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
