package types

// locking module event types
const (
	EventTypeCreateLockedDelegation          = "create_locked_delegation"
	EventTypeLockedDelegationRedelegate      = "locked_delegation_redelegate"
	EventTypeWithdrawLockedDelegationRewards = "withdraw_Locked_delegation_rewards"
	EventTypeToggleAutoRenew                 = "toggle_auto_renew"

	AttributeKeyAutoRenew = "auto_renew"
	AttributeKeyUnlockOn  = "unlock_on"
	AttributeKeyValidator = "validator"
	AttributeKeyDelegator = "delegator"
)
