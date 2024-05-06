package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Wrapper struct
type DistributionHooks struct {
	k Keeper
}

// WARNING: THIS HOOK MUST BE ADDED TO THE COSMOS-SDK FOR MODULE TO WORK!
var _ distributiontypes.DistributionHooks = DistributionHooks{}

// Create new distribution hooks
func (k Keeper) DistributionHooks() DistributionHooks {
	return DistributionHooks{k}
}

// AfterWithdrawDelegationRewards implements types.DistributionHook
//
// This hook is important for a few reasons like keeping the locked rewards and normal rewards in sync
// There's two important moments that are handled by the hook:
//   - Upon the creation of a new locked delegation we also create a delegation
//     The staking hook for delegation will call the rewards hook, calling this hook
//   - Upon a expiration of a locked delegation it will undelegate
//     The staking hook for undelegate will also call this hook
//
// This means that we are making a withdraw at the locked delegation state critical moments
func (h DistributionHooks) AfterWithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, rewards sdk.Coins) error {
	_, err := h.k.withdrawLockedDelegationRewards(ctx, delAddr, valAddr, rewards)
	if err != nil {
		return err
	}
	return nil
}
