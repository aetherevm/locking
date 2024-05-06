package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aetherevm/locking/locking/types"
)

const (
	// Validation Errors
	ErrEmptyRequest   = "empty request"
	ErrEmptyDelegator = "delegator address cannot be empty"
	ErrEmptyValidator = "validator address cannot be empty"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// LockedDelegation queries locked delegatation info for given validator delegator pair
func (k Keeper) LockedDelegations(c context.Context, req *types.QueryLockedDelegationRequest) (*types.QueryLockedDelegationResponse, error) {
	// Validate the parameters
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, ErrEmptyRequest)
	}
	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyDelegator)
	}
	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyValidator)
	}

	// Get the delegator address
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, err
	}

	// Get the validator address
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	// Get the prefix store
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.GetLockedDelegationKey(delAddr, valAddr))

	// Now iterates and get the locked delegations based on pagination
	var lockedDelegations []types.LockedDelegationWithTotalShares
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var lockedDelegation types.LockedDelegation
		err := k.cdc.Unmarshal(value, &lockedDelegation)
		if err != nil {
			return err
		}
		lockedDelegations = append(lockedDelegations, types.LockedDelegationWithTotalShares{
			LockedDelegation: lockedDelegation,
			TotalLocked:      lockedDelegation.TotalShares(),
		})
		return nil
	})
	// The iterator may error out
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLockedDelegationResponse{LockedDelegations: lockedDelegations, Pagination: pageRes}, nil
}

// DelegatorLockedDelegations queries all locked delegations of a given delegator address
func (k Keeper) DelegatorLockedDelegations(c context.Context, req *types.QueryDelegatorLockedDelegationsRequest) (*types.QueryDelegatorLockedDelegationsResponse, error) {
	// Validate the parameters
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, ErrEmptyRequest)
	}
	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyDelegator)
	}

	// Get the delegator address
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, err
	}

	// Get the prefix store
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.GetLockedDelegationPerDelegatorKey(delAddr))

	// Now iterates and get the locked delegations based on pagination
	var lockedDelegations []types.LockedDelegationWithTotalShares
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var lockedDelegation types.LockedDelegation
		err := k.cdc.Unmarshal(value, &lockedDelegation)
		if err != nil {
			return err
		}
		lockedDelegations = append(lockedDelegations, types.LockedDelegationWithTotalShares{
			LockedDelegation: lockedDelegation,
			TotalLocked:      lockedDelegation.TotalShares(),
		})
		return nil
	})
	// The iterator may error out
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDelegatorLockedDelegationsResponse{LockedDelegations: lockedDelegations, Pagination: pageRes}, nil
}

// LockedDelegationRewards implements the types.QueryServer
// returns rewards per delegator validator pair
func (k Keeper) LockedDelegationRewards(c context.Context, req *types.QueryLockedDelegationRewardsRequest) (*types.QueryLockedDelegationRewardsResponse, error) {
	// Validate the parameters
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, ErrEmptyRequest)
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyDelegator)
	}

	// Wrap the context
	ctx := sdk.UnwrapSDKContext(c)

	// Find the validator and delegations
	valAdr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	val := k.stakingKeeper.Validator(ctx, valAdr)
	if val == nil {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorExists, req.ValidatorAddress)
	}
	delAdr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	del := k.stakingKeeper.Delegation(ctx, delAdr, valAdr)
	if del == nil {
		return nil, types.ErrNoDelegationExists
	}

	// Calculate the delegation rewards
	endingPeriod := k.distributionKeeper.IncrementValidatorPeriod(ctx, val)
	distributionRewards := k.distributionKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	rewards, _ := distributionRewards.TruncateDecimal()

	// Calculated the locked rewards
	lockedRewards := k.CalculateLockedDelegationRewards(ctx, delAdr, valAdr, rewards)
	return &types.QueryLockedDelegationRewardsResponse{
		DistributionReward: distributionRewards,
		LockingReward:      lockedRewards,
		Total:              distributionRewards.Add(lockedRewards...),
	}, nil
}

// LockedDelegationTotalRewards implements the types.QueryServer
// returns all rewards per delegator
func (k Keeper) LockedDelegationTotalRewards(c context.Context, req *types.QueryLockedDelegationTotalRewardsRequest) (*types.QueryLockedDelegationTotalRewardsResponse, error) {
	// Validate the parameters
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty delegator address")
	}

	// Wrap the context
	ctx := sdk.UnwrapSDKContext(c)

	total := sdk.DecCoins{}
	var delLockedRewards []types.LockedDelegationDelegatorReward

	delAdr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Iterate over the delegations, building the set
	k.stakingKeeper.IterateDelegations(
		ctx, delAdr,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)

			// Calculate the distribution rewards
			endingPeriod := k.distributionKeeper.IncrementValidatorPeriod(ctx, val)
			distributionRewards := k.distributionKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
			delReward, _ := distributionRewards.TruncateDecimal()

			// Calculated the locked rewards
			lockingRewards := k.CalculateLockedDelegationRewards(ctx, delAdr, valAddr, delReward)

			delTotal := distributionRewards.Add(lockingRewards...)
			delLockedRewards = append(delLockedRewards, types.LockedDelegationDelegatorReward{
				ValidatorAddress:   valAddr.String(),
				DistributionReward: distributionRewards,
				LockingReward:      lockingRewards,
				Total:              delTotal,
			})
			total = total.Add(delTotal...)
			return false
		},
	)

	return &types.QueryLockedDelegationTotalRewardsResponse{Rewards: delLockedRewards, Total: total}, nil
}
