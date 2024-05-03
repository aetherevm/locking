// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package locking

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/keeper"
	"github.com/aetherevm/locking/locking/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k *keeper.Keeper,
	data types.GenesisState,
) []abci.ValidatorUpdate {
	// Set the module parameters
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(err)
	}

	// Set the locked delegations and the initial ID for the id counter
	initialID := uint64(0)
	for _, lockedDelegation := range data.LockedDelegations {
		err = k.SetLockedDelegation(ctx, lockedDelegation)
		if err != nil {
			panic(err)
		}
		for _, entry := range lockedDelegation.Entries {
			// Set it to the queue and to the lookup
			k.InsertLockedDelegationQueue(ctx, lockedDelegation, entry.UnlockOn)
			err := k.SetLockedDelegationByEntryID(ctx, lockedDelegation, entry.Id)
			if err != nil {
				panic(err)
			}

			if entry.Id > initialID {
				initialID = entry.Id
			}
		}
	}
	if initialID > 0 {
		k.SetInitialLockedDelegationEntryID(ctx, initialID)
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	// Get the module parameters
	params := k.GetParams(ctx)

	// Get the locked delegations records
	lockedDelegations := k.GetAllLockedDelegations(ctx)

	// Return the genesis state
	return types.NewGenesisState(
		params,
		lockedDelegations,
	)
}
