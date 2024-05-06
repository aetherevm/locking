package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/aetherevm/locking/locking/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreateLockedDelegationCmd(),
		NewRedelegateLockedDelegationsCmd(),
		NewToggleAutoRenewCmd(),
	)

	return cmd
}

// NewCreateLockedDelegationCmd returns a CLI command handler for creating a MsgCreateLockedDelegation transaction.
func NewCreateLockedDelegationCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "create-locked-delegation [validator-addr] [amount] [duration] [auto-renew]",
		Args:  cobra.RangeArgs(3, 4),
		Short: "Delegates and creates a Locked Delegation to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate and create a Locked delegation with an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx locking create-locked-delegation %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake 123123s --auto-renew=true --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// Get the addresses
			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Parse the amount
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			// Parse the lock duration
			lockDuration, err := time.ParseDuration(args[2])
			if err != nil {
				return err
			}

			// Parse the auto-renew flag with default value as true
			var autoRenew bool
			if len(args) == 4 {
				autoRenew, err = strconv.ParseBool(args[3])
				if err != nil {
					return err
				}
			} else {
				autoRenew, err = cmd.Flags().GetBool("auto-renew")
				if err != nil {
					return err
				}
			}

			// Create the message
			msg := types.NewMsgCreateLockedDelegation(
				delAddr,
				valAddr,
				amount,
				lockDuration,
				autoRenew,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Broadcast
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	// Add auto-renew flag, it is optional and by default true
	cmd.Flags().Bool("auto-renew", true, "Automatically renew the locked delegation when it expires")

	return cmd
}

// NewRedelegateLockedDelegationsCmd returns a CLI command handler for creating a MsgRedelegateLockedDelegations transaction
func NewRedelegateLockedDelegationsCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegate-locked-delegations [src-validator-addr] [dst-validator-addr] [ids]",
		Short: "Redelegate locked tokens from one validator to another",
		Args:  cobra.RangeArgs(2, 3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate and move locked delegations from one validator to another.
Specific IDs can be provided as a comma separated list. 
If no ID is provided, the redelegation will happen to all locked delegation entries.

Example:
$ %s tx locking redelegate-locked-delegations %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1,2,3 --from mykey
`,
				version.AppName, bech32PrefixValAddr, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()
			// Parse the src validator address
			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Parse the dst validator address
			valDstAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			// Parse the ID list
			var ids []uint64
			if len(args) > 2 {
				for _, idStr := range strings.Split(args[2], ",") {
					id, err := strconv.ParseUint(idStr, 10, 64)
					if err != nil {
						return err
					}
					ids = append(ids, id)
				}
			}

			msg := types.NewMsgRedelegateLockedDelegations(delAddr, valSrcAddr, valDstAddr, ids)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewToggleAutoRenewCmd returns a CLI command handler for creating a MsgToggleAutoRenew transaction
func NewToggleAutoRenewCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "toggle-auto-renew [validator-addr] [entry-id]",
		Short: "Toggle a auto renew flag on a locked delegation entry",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Toggle a auto renew flag on a locked delegation entry

Example:
$ %s tx locking toggle-auto-renew %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 123 --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// Parse the address
			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Parse the ID
			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// Generate the message
			msg := types.NewMsgToggleAutoRenew(
				delAddr,
				valAddr,
				id,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Broadcast
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
