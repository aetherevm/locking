package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aetherevm/locking/locking/types"
)

// Errors for the keeper
const (
	ErrAuthorityNotAccount = "authority is not a valid acc address"
)

// Locking module keeper
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	stakingKeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper
	bankKeeper         types.BankKeeper

	authority string
}

// NewKeeper creates new instances of the locking Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.Codec,
	sk types.StakingKeeper,
	dk types.DistributionKeeper,
	bk types.BankKeeper,
	authority string,
) *Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(ErrAuthorityNotAccount)
	}

	return &Keeper{
		storeKey:           storeKey,
		cdc:                cdc,
		stakingKeeper:      sk,
		distributionKeeper: dk,
		bankKeeper:         bk,
		authority:          authority,
	}
}

// SetStakingKeeper is a solver for a issue where keeper doesn't carry the hooks
func (k *Keeper) SetStakingKeeper(sk types.StakingKeeper) {
	k.stakingKeeper = sk
}

// SetDistributionKeeper is a solver for a issue where keeper doesn't carry the hooks
func (k *Keeper) SetDistributionKeeper(dk types.DistributionKeeper) {
	k.distributionKeeper = dk
}

// GetAuthority returns the x/locking module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}
