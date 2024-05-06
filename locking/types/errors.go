package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrFailedUnmarshalLockedDelegation        = errorsmod.Register(ModuleName, 1, "failed to unmarshal locked delegation")
	ErrCreateLockedDelegationDurationUnmatch  = errorsmod.Register(ModuleName, 2, "locked delegation does not have a corresponding rate duration in params")
	ErrMaxLockedDelegationEntriesReached      = errorsmod.Register(ModuleName, 3, "the max number of entries for the current locked delegation has been reached")
	ErrLDRedelegationMaxEntriesReached        = errorsmod.Register(ModuleName, 4, "this redelegation will reach the max entries for the target locked delegation")
	ErrNoLockedSharesToRedelegate             = errorsmod.Register(ModuleName, 5, "no locked shares were found to redelegate")
	ErrUndelegateUnbondAmountDiffer           = errorsmod.Register(ModuleName, 6, "undelegate validate unbound amount differ from required amount")
	ErrLockedDelegationNotFound               = errorsmod.Register(ModuleName, 7, "locked delegation for delegator and validator addresses pair not found")
	ErrLockedDelegationRedelegationZeroShares = errorsmod.Register(ModuleName, 8, "can't redelegate a locked delegation with zero shares")
	ErrLockedDelegationEntryNotFound          = errorsmod.Register(ModuleName, 9, "locked delegation entry for specified id not found")
	ErrLockedSharesSmallerThanDelegation      = errorsmod.Register(ModuleName, 10, "total locked shares should not be smaller than delegated shares")
	ErrRedelegationIdsBiggerThanMaxEntries    = errorsmod.Register(ModuleName, 11, "requested redelegation ids list length is bigger than max entries")
	ErrNoValidatorExists                      = errorsmod.Register(ModuleName, 12, "validator does not exist")
	ErrNoDelegationExists                     = errorsmod.Register(ModuleName, 13, "delegation does not exist")
)
