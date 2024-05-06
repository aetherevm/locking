package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/types"
)

// CalculateLockedDelegationRewards calculates the locked delegation rewards for a validator
// we use a ratio from all the locked delegation weights and the delegation shares
func (k Keeper) CalculateLockedDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, rewards sdk.Coins) sdk.DecCoins {
	// Return if rewards is zero
	if rewards.IsZero() {
		return sdk.NewDecCoins()
	}

	// Fetch the normal distribution rewards
	delegation, found := k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return sdk.NewDecCoins()
	}
	// Fetch the locked delegation
	lockedDelegation, found := k.GetLockedDelegation(ctx, delAddr, valAddr)
	if !found {
		return sdk.NewDecCoins()
	}

	// Calculate the reward rate by the entries and the delegation
	ratio := lockedDelegation.CalculateDelegationRatio(delegation.Shares)

	// Convert the tokens to DecCoins and apply the ratio
	rewardsDecCoins := sdk.NewDecCoinsFromCoins(rewards...)
	return rewardsDecCoins.MulDecTruncate(ratio)
}

// withdrawLockedDelegationRewards does the minting of new coins on top of delegation rewards withdraw
// currently we are minting the new rewards
func (k Keeper) withdrawLockedDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, rewards sdk.Coins) (sdk.Coins, error) {
	// Calculate the rewards on top of the normal delegation rewards
	rewardsRaw := k.CalculateLockedDelegationRewards(ctx, delAddr, valAddr, rewards)

	// Truncate reward dec coins, we don't care about remainder at this point
	// this also converts the DecCoins to Coins
	finalRewards, _ := rewardsRaw.TruncateDecimal()

	// if the rewards is not zero, we are safe to mint and send the rewards to the delegator
	if !finalRewards.IsZero() {
		// Mint new tokens
		err := k.bankKeeper.MintCoins(ctx, types.ModuleName, finalRewards)
		if err != nil {
			return nil, err
		}
		// Send to the delegator address
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delAddr, finalRewards)
		if err != nil {
			return nil, err
		}
	}

	// Emit the rewards collection event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawLockedDelegationRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
		),
	)

	return finalRewards, nil
}
