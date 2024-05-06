package keeper

import (
	"context"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrorstypes "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/aetherevm/locking/locking/types"
)

const (
	ErrInvalidDenom     = "invalid coin denomination: got %s, expected %s"
	ErrInvalidAuthority = "invalid authority; expected %s, got %s"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the locking MsgServer interface
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateLockedDelegation creates a new locked delegation
// But before it creates a new normal delegation and throw the new locked delegation on top
// This is built on top of the staking delegate msg server implementation
func (ms msgServer) CreateLockedDelegation(goCtx context.Context, msg *types.MsgCreateLockedDelegation) (*types.MsgCreateLockedDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the validator and delegator address
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Check the input msg denomination
	bondDenom := ms.stakingKeeper.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrorstypes.ErrInvalidRequest, ErrInvalidDenom, msg.Amount.Denom, bondDenom,
		)
	}

	// Create a new locked delegation entry and a staking delegation
	entry, err := ms.Keeper.CreateLockedDelegationEntryAndDelegate(
		ctx,
		delAddr,
		valAddr,
		msg.Amount.Amount,
		msg.LockDuration,
		msg.AutoRenew,
	)
	if err != nil {
		return nil, err
	}

	// Do the telemetry for the staking type
	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, stakingtypes.ModuleName, "delegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	// Emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateLockedDelegation,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, entry.Shares.String()),
			sdk.NewAttribute(types.AttributeKeyUnlockOn, entry.UnlockOn.String()),
			sdk.NewAttribute(types.AttributeKeyAutoRenew, strconv.FormatBool(entry.AutoRenew)),
		),
	})

	return &types.MsgCreateLockedDelegationResponse{}, nil
}

// RedelegateLockedDelegations creates a redelegates tokens
// But before we do the normal undelegation
// This is built on top of the staking BeginRedelegate msg server implementation
func (ms msgServer) RedelegateLockedDelegations(goCtx context.Context, msg *types.MsgRedelegateLockedDelegations) (*types.MsgRedelegateLockedDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the number of redelegate ids is bigger than the max entries
	if uint32(len(msg.Ids)) > ms.Keeper.MaxEntries(ctx) {
		return nil, types.ErrRedelegationIdsBiggerThanMaxEntries
	}

	// Get the src and dst validator addresses and delegator address
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, err
	}
	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}

	// Now we can do the locked delegation redelegation
	sharesRedelegated, tokensRedelegated, err := ms.Keeper.LockedDelegationAndStakingRedelegation(
		ctx,
		delAddr,
		valSrcAddr,
		valDstAddr,
		msg.Ids,
	)
	if err != nil {
		return nil, err
	}

	// Do the correct telemetry
	bondDenom := ms.stakingKeeper.BondDenom(ctx)
	if tokensRedelegated.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "redelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(tokensRedelegated.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", bondDenom)},
			)
		}()
	}

	// Emit the events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLockedDelegationRedelegate,
			sdk.NewAttribute(stakingtypes.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(stakingtypes.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sharesRedelegated.String()),
		),
	})

	return &types.MsgRedelegateLockedDelegationsResponse{}, nil
}

// ToggleAutoRenew toggles a auto renew for a single locked delegation entry
func (ms msgServer) ToggleAutoRenew(goCtx context.Context, msg *types.MsgToggleAutoRenew) (*types.MsgToggleAutoRenewResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the addresses
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Update the entry
	entry, err := ms.Keeper.ToggleLockedDelegationEntryAutoRenew(ctx, delAddr, valAddr, msg.Id)
	if err != nil {
		return nil, err
	}

	// Emit the events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeToggleAutoRenew,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyAutoRenew, strconv.FormatBool(entry.AutoRenew)),
		),
	})

	return &types.MsgToggleAutoRenewResponse{}, nil
}

// UpdateParams updates params though a proposal
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if ms.authority != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, ErrInvalidAuthority, ms.authority, msg.Authority)
	}

	// Update params
	if err := ms.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
