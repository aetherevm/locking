// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/aetherevm/locking/locking/types"
)

// TODO: Test this file

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group claim queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(GetCmdQueryParams())
	cmd.AddCommand(GetCmdQueryLockedDelegationsTo())
	cmd.AddCommand(GetCmdQueryLockedDelegations())
	cmd.AddCommand(GetCmdQueryDelegatorRewards())
	return cmd
}

// GetCmdQueryParams implements a command to return the current locking params
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current locking parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLockedDelegations implements the command to query locked delegations delegations
// for an individual delegator on an individual validator
func GetCmdQueryLockedDelegationsTo() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "locked-delegations-to [delegator-addr] [validator-addr]",
		Short: "Query locked delegations based on address and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query locked delegations for an individual delegator on an individual validator.

Example:
$ %s query locking locked-delegations-to %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryLockedDelegationRequest{
				DelegatorAddr: delAddr.String(),
				ValidatorAddr: valAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.LockedDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "locked-delegations")

	return cmd
}

// GetCmdQueryLockedDelegations implements the command to query locked delegations delegations
// for an individual delegator
func GetCmdQueryLockedDelegations() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "locked-delegations [delegator-addr]",
		Short: "Query locked delegations based on address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query locked delegations for an individual delegator.

Example:
$ %s query locking locked-delegations %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorLockedDelegationsRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.DelegatorLockedDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "locked-delegations")

	return cmd
}

// GetCmdQueryDelegatorRewards implements the query delegator rewards command
func GetCmdQueryDelegatorRewards() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "rewards [delegator-addr] [validator-addr]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query all locked delegations rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all locked delegation rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
$ %s query locking rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query locking rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// query for rewards from a particular delegation
			ctx := cmd.Context()
			// For validator delegator pair
			if len(args) == 2 {
				validatorAddr, err := sdk.ValAddressFromBech32(args[1])
				if err != nil {
					return err
				}

				res, err := queryClient.LockedDelegationRewards(
					ctx,
					&types.QueryLockedDelegationRewardsRequest{
						DelegatorAddress: delegatorAddr.String(),
						ValidatorAddress: validatorAddr.String(),
					},
				)
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			// For all validators
			res, err := queryClient.LockedDelegationTotalRewards(
				ctx,
				&types.QueryLockedDelegationTotalRewardsRequest{DelegatorAddress: delegatorAddr.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
