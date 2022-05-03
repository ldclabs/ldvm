// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

var LocalGenesisConfigJSON = `
{
  "chain": {
    "chainID": 2357,
    "maxTotalSupply": 1000000000000000000,
    "message": "Hello, LDVM!",
    "feeConfig": {
      "startHeight": 0,
      "thresholdGas": 1000,
      "minGasPrice": 10000,
      "maxGasPrice": 100000,
      "gasRebateRate": 1000,
      "maxTxGas": 42000000,
      "maxBlockTxsSize": 4200000,
      "minTokenPledge": 10000000000000,
      "minValidatorStake": 1000000000000,
      "maxValidatorStake": 100000000000000,
      "minDelegatorStake": 100000000000,
      "minWithdrawFee": 20000
    }
  },
  "alloc": {
    "0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF": {
      "balance": 400000000000000000,
      "threshold": 2,
      "keepers": [
        "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
        "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"
      ]
    },
    "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC": {
      "balance": 100000000000000000,
      "threshold": 1,
      "keepers": [
        "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
        "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"
      ]
    }
  }
}
`
