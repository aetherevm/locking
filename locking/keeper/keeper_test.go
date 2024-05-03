// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package keeper_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/aetherevm/locking/locking/keeper"
	"github.com/aetherevm/locking/locking/types"
	"github.com/aetherevm/locking/testing/simapp"
)

var PowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))

// big time used on testing
var (
	bigTime  = time.Unix(1<<63-1, 0)
	authAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
)

var (
	TestChainId = "chain-0"
)

// KeeperTestSuite is a suite of unit tests
type KeeperTestSuite struct {
	suite.Suite
	ctx     sdk.Context
	app     *simapp.SimApp
	k       *keeper.Keeper
	msgSrvr types.MsgServer
}

// SetupTest sets up the testing environment before each test, initializing the app, context
func (suite *KeeperTestSuite) SetupTest() {
	suite.app = simapp.Setup(suite.T(), false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 2, ChainID: TestChainId, Time: time.Now().UTC()})
	startTime := time.Now().UTC()
	suite.ctx = suite.ctx.WithBlockTime(startTime)

	// Set the locking keeper
	suite.k = suite.app.LockingKeeper

	// Set the msg server
	suite.msgSrvr = keeper.NewMsgServerImpl(suite.k)

	// Mint tokens to the module
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.TokensFromConsensusPower(100_000_000_000, sdk.DefaultPowerReduction)),
	))
	suite.Require().NoError(err)

	// Send tokens to the distribution module
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(
		suite.ctx,
		types.ModuleName,
		distributiontypes.ModuleName,
		sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.TokensFromConsensusPower(50_000_000_000, sdk.DefaultPowerReduction))),
	)

	suite.Require().NoError(err)
}

// TestKeeperTestSuite runs all the tests in the KeeperTestSuite.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
