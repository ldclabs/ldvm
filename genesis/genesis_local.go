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
      "maxTxGas": 42000000,
      "maxBlockTxsSize": 4200000,
      "gasRebateRate": 1000,
      "minTokenPledge": 10000000000000,
      "minStakePledge": 1000000000000
    }
  },
  "alloc": {
    "0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff": {
      "balance": 400000000000000000,
      "threshold": 2,
      "keepers": [
        "jbl8fOziScK5i9wCJsxMKle_UvwKxwPH",
        "RBccN_9de3u43K1cgfFihKIp5kE1lmGG",
        "OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"
      ]
    },
    "0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc": {
      "balance": 100000000000000000,
      "threshold": 1,
      "keepers": [
        "jbl8fOziScK5i9wCJsxMKle_UvwKxwPH",
        "RBccN_9de3u43K1cgfFihKIp5kE1lmGG",
        "vlU2vQ63fTjirRWtJJvVcncVnyBaFovseXYCgCQkS-avAHc4"
      ]
    }
  }
}
`
