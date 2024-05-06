package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock iterates over the locked delegations, unlocking the ones that has been expired
func (k Keeper) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	currTime := ctx.BlockTime()

	// Dequeue locked delegation pairs
	expiredPairs := k.DequeueExpiredLockedDelegations(ctx, currTime)

	// Complete the expired entries
	for _, pair := range expiredPairs {
		err := k.CompleteLockedDelegations(ctx, pair)
		// There's no problem into panicking at this point
		if err != nil {
			panic(err)
		}
	}
	// Returns a empty validator set to complete the endblock interface
	return []abci.ValidatorUpdate{}
}
