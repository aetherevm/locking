package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/types"
)

// MaxEntries - returns the max entries a val and del pair can hold
func (k Keeper) MaxEntries(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxEntries
}

// Rates - Returns selected rates for rewards generation
func (k Keeper) Rates(ctx sdk.Context) []types.Rate {
	return k.GetParams(ctx).Rates
}

// GetParams returns the total set of claim parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets claim parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	err := params.Validate()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}
