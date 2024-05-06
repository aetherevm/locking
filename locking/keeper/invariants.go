package keeper

import (
	"fmt"

	"github.com/aetherevm/locking/locking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	InvariantLDZero             = "\tlocked delegation with zero shares: %+v\n"
	InvariantLDNoDelegation     = "\tlocked delegation with no delegation: %+v\n"
	InvariantLDBiggerDelegation = "\tlocked delegation with shares bigger than delegation: %+v\n"

	InvariantLDFound = "%d invalid locked delegations found\n%s"
)

// RegisterInvariants registers all locking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "valid-locked-delegation",
		ValidLockedDelegation(k))
}

// AllInvariants runs all invariants of the locking module.
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return ValidLockedDelegation(k)(ctx)
	}
}

// ValidLockedDelegation checks if all locked delegation has a smaller share than the corresponding delegation
func ValidLockedDelegation(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		lockedDelegations := k.GetAllLockedDelegations(ctx)
		for _, lockedDelegation := range lockedDelegations {
			if lockedDelegation.TotalShares().IsZero() {
				count++
				msg += fmt.Sprintf(InvariantLDZero, lockedDelegation)
			}

			// We skip the error, if it exists no delegation will be found
			valAddr, _ := lockedDelegation.GetValidatorAddr()

			delegation, found := k.stakingKeeper.GetDelegation(ctx, lockedDelegation.GetDelegatorAddr(), valAddr)
			if !found {
				count++
				msg += fmt.Sprintf(InvariantLDNoDelegation, lockedDelegation)
			}
			if found && delegation.Shares.LT(lockedDelegation.TotalShares()) {
				count++
				msg += fmt.Sprintf(InvariantLDBiggerDelegation, lockedDelegation)
			}

		}

		broken := count != 0

		return sdk.FormatInvariant(types.ModuleName, "valid locked delegations", fmt.Sprintf(
			InvariantLDFound, count, msg)), broken
	}
}
