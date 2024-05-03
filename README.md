# Locking / Incentivized Staking

# Concept

The Locking module in Cosmos-SDK facilitates the establishment of delegated asset locking, aimed at rewarding delegators based on the duration of their lock.
By engaging in a locking agreement, delegators can earn additional rewards, with the reward rate increasing with the length of the lock period.

The locking is always cleared when it expires. Two states are possible at this point:

- Renewable: The locking gets renewed after expiration
- Not renewable: An undelegation is created after expiration

To avoid spamming the system, each locking delegation can have only a set of entries as defined in `genesis.json`.

## Rewards

Delegations that get locked reward the user with specific rates.

Each locking can only be created on top of a rate that exists as a state. Each rate is formed by a duration and a reward ratio and rates may take this form:

- 08 months; 2.2% reward
- 16 months; 3.3% reward
- 24 months; 4.4% reward
- 32 months; 5.5% reward

These reward rates are applied on top of the normal delegation rewards, forming a multiplier for the final rewards, as shown in the following formula:

$$WeightedRatio=\frac{\sum_{i=1}^n (𝑒𝑛𝑡𝑟𝑦𝑖.𝑠ℎ𝑎𝑟𝑒𝑠 * 𝑒𝑛𝑡𝑟𝑦𝑖.𝑟𝑎𝑡𝑒)} {TotalShares}$$

Where:

- N: Number of entries on a Locked Delegation
- Entry: Each one of the possible locked delegation entries
- TotalShares: Sum of all the locked shares

$$DelegationRatio=\frac{TotalShares} {DelegatedShares} * WeightedRatio$$

Where:

- DelegationRatio: The final multiplier for the rewards
- TotalShares: Sum of all the locked shares
- DelegatedShares: Total shares delegated (from the staking module)

Here, examples of possible rewards are tabulated:

| Distribution Reward | Delegation Shares | LDE1\*         | LDE2\*         | Final Ratio | Locking reward | Final Reward |
| ------------------- | ----------------- | -------------- | -------------- | ----------- | -------------- | ------------ |
| 10_000evm           | 100_000evm        | 50_000evm - 5% | 50_000evm - 5% | 0.05        | 500evm         | 10_500evm    |
| 50_000evm           | 100_000evm        | 25_000evm - 5% | 25_000evm - 5% | 0.025       | 1_250evm       | 51_250evm    |
| 200_000evm          | 250_000evm        | 50_000evm - 5% | 20_000evm - 4% | 0.0132      | 2_640evm       | 202_640evm   |

- LD is a locked delegation entry (Amount - Ratio)
- Further examples can be found at: [locked_delegation_test](./types/locked_delegation_test.go)

Rewards are collected using the distribution module and can be collected at any point.
New rewards are minted directly through the bank module to the user account.

## Key features

- **Creation of Locking Delegations**: Enable users to create locked delegations to earn additional rewards.
- **Renewable and Non-renewable Locks**: Specify the behavior of locks upon expiration - whether they get renewed or lead to undelegation.
- **Integration with Distribution Module**: Allow users to claim rewards using the distribution module at any time.

# State

The module's state is saved on Chain and comprises two main components:

- Params
- LockedDelegations

## Params

This state stores the parameters of the locking module. The structure is defined as follows:

- Maximum Entries: Define the maximum entries for locked delegation per pair
- Reward Rates: List the reward rates for different lock durations

```proto
// Params defines the locking module's parameters.
message Params {
  option (gogoproto.goproto_stringer) = false;
  // max_entries is the max entries for locked delegation (per pair).
  uint32 max_entries = 1;
  // Rates are the rates of rewards
  repeated Rate rates = 2 [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
}

// Rate are the rate of rewards for the locked delegations
message Rate {
  option (gogoproto.equal) = true;
  // Duration is the lock duration
  google.protobuf.Duration duration = 1 [(gogoproto.nullable) = false, (gogoproto.stdduration) = true];
  // The rate used on calculation
  string rate = 2 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

## LockedDelegations

This state contains a set of locked delegations, and it's defined as follows:

- **Delegator and Validator Addresses**: Store the addresses of the delegator and validator involved in the locked delegation.
- **Locked Delegation Entries**: Maintain a record of all locked delegation entries, including shares, rate, unlock time, and auto-renew preference.

```proto
// LockedDelegation defines the locking locked delegations
message LockedDelegation {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_getters)  = false;
  option (gogoproto.goproto_stringer) = false;

  // delegator_address is the bech32-encoded address of the delegator
  string delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // validator_address is the bech32-encoded address of the validator
  string validator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // entries are all the lockings made on top of the pair
  repeated LockedDelegationEntry entries = 3
  [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
}

message LockedDelegationEntry {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;
  // shares define the locked shares
  string shares = 1 [
    (cosmos_proto.scalar)  = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable)   = false
  ];
  // The rate used on calculation
  Rate rate = 2 [
    (gogoproto.nullable) = false,
    (amino.dont_omitempty) = true
  ];
  // unlock_on defines when the delegation will be unlocked
  google.protobuf.Timestamp unlock_on = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.stdtime) = true
  ];
  // auto_renew defines if the delegator wants to auto renew the locking after expiration
  bool auto_renew = 4 [
    (gogoproto.moretags) = "yaml:\"undelegate\""
  ];
}
```

# Messages

In this section, we describe the processing of the locking messages and the corresponding updates to the state.

## CreateLockedDelegation

This message creates the following:

- A new delegation
- A locked delegation on top the delegation

Here's the definition of the message:

```
// Msg defines the locking Msg service.
service Msg {
    // CreateLockedDelegation defines a method for creating a new locked delegation.
    rpc CreateLockedDelegation(MsgCreateLockedDelegation) returns (MsgCreateLockedDelegationResponse);
}

// MsgCreateLockedDelegation defines a SDK message for creating a locked delegation
message MsgCreateLockedDelegation {
    option (cosmos.msg.v1.signer) = "delegator_address";
    option (amino.name)           = "aether/CreateLockedDelegation";

    option (gogoproto.equal)           = false;
    option (gogoproto.goproto_getters) = false;

    // delegator_address is the target delegator address
    string                   delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
    // validator_address is the target validator address
    string                   validator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
    // validator_address is the amount that will delegated and locked
    cosmos.base.v1beta1.Coin amount            = 3 [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
    // lock_duration is for how long the locking will last
    google.protobuf.Duration lock_duration     = 4 [(gogoproto.nullable) = false, (gogoproto.stdduration) = true];
    // undelegate is if the delegator wants to auto undelegate after the locking is over
    bool                     undelegate        = 5;
}

// MsgCreateLockedDelegationResponse defines the Msg/CreateLockedDelegation response type.
message MsgCreateLockedDelegationResponse {}
```

This message will fail under the following conditions:

- If the delegation creation fails
- If the user has reached the maximum number of entries for that validator

Upon successful processing:

- A new delegation is created
- A new entry is added to the locking delegation object

## RedelegateLockedDelegations

This message does the following:

- Redelegates all the locked delegation values from one validator to another
- Move all locked delegation entries from one validator to another

Here's the definition of the message:

```
// Msg defines the locking Msg service.
service Msg {
    // RedelegateLockedDelegation defines a method for performing a redelegation
    // of locked delegations
    rpc RedelegateLockedDelegations(MsgRedelegateLockedDelegations) returns (MsgRedelegateLockedDelegationsResponse);
}

// MsgRedelegateLockedDelegation defines a SDK message for performing a redelegation
// of locked delegations
message MsgRedelegateLockedDelegations {
    option (cosmos.msg.v1.signer) = "delegator_address";
    option (amino.name)           = "cosmos-sdk/MsgRedelegateLockedDelegation";

    option (gogoproto.equal)           = false;
    option (gogoproto.goproto_getters) = false;

    string                   delegator_address     = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
    string                   validator_src_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
    string                   validator_dst_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

// MsgRedelegateLockedDelegationsResponse defines the Msg/MsgRedelegateLockedDelegation response type.
message MsgRedelegateLockedDelegationsResponse{
google.protobuf.Timestamp completion_time = 1
    [(gogoproto.nullable) = false, (amino.dont_omitempty) = true, (gogoproto.stdtime) = true];
}
```

This message will fail under the following conditions:

- If the redelegation fails

Upon successful processing:

- A relegation is done between the source and destination validator
- The locked delegation entries are moved from the source to the destination validator

# End-Block

At the end of each block, Aether checks for expired locked delegations. The following is done:

- We iterate over a queue of validator delegator pairs
  - New items are added to the queue at each locked delegation entry creation
- For each expired locked delegation entry:
  - The item is removed from the queue
  - The item is removed from the locked delegation entries list
  - If renew is enabled:
    - The entry is updated with a new unlock at the last unlock time + original rate duration
  - If the entry isn't renewable:
    - A undelegation is created

This whole process ensures that at the end of each block, we only iterate over expired entries.

# Events

The claim module emits the following events:

# Withdraw locked delegation rewards

| Type     | Attribute Key                      | Attribute Value                                |
| -------- | ---------------------------------- | ---------------------------------------------- |
| withdraw | withdraw_Locked_delegation_rewards | {reward, validator address, delegator address} |

# Msg's

## CreateLockedDelegation

| Type                     | Attribute Key            | Attribute Value                            |
| ------------------------ | ------------------------ | ------------------------------------------ |
| create locked delegation | create_locked_delegation | {validator, shares, unlock on, auto renew} |

## RedelegateLockedDelegations

| Type                     | Attribute Key                | Attribute Value                                          |
| ------------------------ | ---------------------------- | -------------------------------------------------------- |
| create locked delegation | locked_delegation_redelegate | {validator source, validator destination, locked shares} |
