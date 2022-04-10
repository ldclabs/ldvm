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
      "maxBlockSize": 4200000,
      "maxBlockMiners": 128,
      "minMinerStake": 1000000000000
    },
    "feeConfigThreshold": 2,
    "feeConfigKeepers": [
      "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
      "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"
    ],
    "nameServiceThreshold": 2,
    "nameServiceKeepers": [
      "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
      "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"
    ]
  },
  "block": {
    "gasRebateRate": 500,
    "gasPrice": 1000
  },
  "alloc": {
    "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC": {
      "balance": 500000000000000000,
      "threshold": 1,
      "keepers": [
        "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
        "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"
      ]
    }
  }
}
`
