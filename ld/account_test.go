// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStakeConfig(t *testing.T) {
	assert := assert.New(t)

	var cfg *StakeConfig
	assert.ErrorContains(cfg.SyntacticVerify(), "nil pointer")

	cfg = &StakeConfig{Token: ids.TokenSymbol{1, 2, 3}}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid token")

	cfg = &StakeConfig{Type: 3}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid type")

	cfg = &StakeConfig{WithdrawFee: 0}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid withdrawFee")
	cfg = &StakeConfig{WithdrawFee: 200_001}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid withdrawFee")

	cfg = &StakeConfig{WithdrawFee: 1, MinAmount: new(big.Int)}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid minAmount")
	cfg = &StakeConfig{WithdrawFee: 1, MinAmount: new(big.Int).SetUint64(100), MaxAmount: new(big.Int).SetUint64(99)}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid maxAmount")

	cfg = &StakeConfig{
		WithdrawFee: 1,
		MinAmount:   new(big.Int).SetUint64(100),
		MaxAmount:   new(big.Int).SetUint64(100),
	}
	assert.NoError(cfg.SyntacticVerify())
	cbordata, err := cfg.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(cfg)
	require.NoError(t, err)

	assert.Equal(`{"token":"","type":0,"lockTime":0,"withdrawFee":1,"minAmount":100,"maxAmount":100}`,
		string(jsondata))

	cfg2 := &StakeConfig{}
	assert.NoError(cfg2.Unmarshal(cbordata))
	assert.NoError(cfg2.SyntacticVerify())

	cbordata2, err := cfg2.Marshal()
	require.NoError(t, err)
	jsondata2, err := json.Marshal(cfg2)
	require.NoError(t, err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}

func TestLendingConfig(t *testing.T) {
	assert := assert.New(t)

	var cfg *LendingConfig
	assert.ErrorContains(cfg.SyntacticVerify(), "nil pointer")

	cfg = &LendingConfig{Token: ids.TokenSymbol{1, 2, 3}}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid token")

	cfg = &LendingConfig{DailyInterest: 0}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid dailyInterest")
	cfg = &LendingConfig{DailyInterest: 10_001}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid dailyInterest")

	cfg = &LendingConfig{DailyInterest: 1, OverdueInterest: 0}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid overdueInterest")
	cfg = &LendingConfig{DailyInterest: 1, OverdueInterest: 10_001}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid overdueInterest")

	cfg = &LendingConfig{DailyInterest: 1, OverdueInterest: 1, MinAmount: new(big.Int)}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid minAmount")
	cfg = &LendingConfig{DailyInterest: 1, OverdueInterest: 1,
		MinAmount: new(big.Int).SetUint64(100), MaxAmount: new(big.Int).SetUint64(99)}
	assert.ErrorContains(cfg.SyntacticVerify(), "invalid maxAmount")

	token, _ := ids.TokenFromStr("$LDC")
	cfg = &LendingConfig{
		Token:           token,
		DailyInterest:   10,
		OverdueInterest: 1,
		MinAmount:       new(big.Int).SetUint64(100),
		MaxAmount:       new(big.Int).SetUint64(100),
	}
	assert.NoError(cfg.SyntacticVerify())
	cbordata, err := cfg.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(cfg)
	require.NoError(t, err)

	assert.Equal(`{"token":"$LDC","dailyInterest":10,"overdueInterest":1,"minAmount":100,"maxAmount":100}`,
		string(jsondata))

	cfg2 := &LendingConfig{}
	assert.NoError(cfg2.Unmarshal(cbordata))
	assert.NoError(cfg2.SyntacticVerify())

	cbordata2, err := cfg2.Marshal()
	require.NoError(t, err)
	jsondata2, err := json.Marshal(cfg2)
	require.NoError(t, err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}

func TestAccount(t *testing.T) {
	assert := assert.New(t)

	var acc *Account
	assert.ErrorContains(acc.SyntacticVerify(), "nil pointer")

	acc = &Account{}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid balance")

	acc = &Account{Balance: big.NewInt(-1)}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid balance")

	acc = &Account{Balance: big.NewInt(0), Threshold: 0}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid keepers")

	acc = &Account{Balance: big.NewInt(0), Threshold: 1, Keepers: signer.Keys{}}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid threshold")

	acc = &Account{Balance: big.NewInt(0), Keepers: signer.Keys{}}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid tokens")

	acc = &Account{
		Balance: big.NewInt(0),
		Keepers: signer.Keys{},
		Tokens:  make(map[string]*big.Int),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid nonceTable")

	acc = &Account{
		Type:       3,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid type")

	acc = &Account{
		Type:           NativeAccount,
		Balance:        big.NewInt(0),
		Keepers:        signer.Keys{},
		Tokens:         make(map[string]*big.Int),
		NonceTable:     make(map[uint64][]uint64),
		MaxTotalSupply: big.NewInt(0),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid maxTotalSupply, should be nil")

	acc = &Account{
		Type:       NativeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Approver:   signer.Key{},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid approver")

	acc = &Account{
		Type:       NativeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Approver:   signer.Key(ids.EmptyID20[:]),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid approver")

	acc = &Account{
		Type:       NativeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Approver:   signer.Key(ids.EmptyID32[:]),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid approver")

	acc = &Account{
		Type:        NativeAccount,
		Balance:     big.NewInt(0),
		Keepers:     signer.Keys{},
		Tokens:      make(map[string]*big.Int),
		NonceTable:  make(map[uint64][]uint64),
		ApproveList: TxTypes{TxType(255)},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid TxType TypeUnknown(255) in approveList")

	acc = &Account{
		Type:        NativeAccount,
		Balance:     big.NewInt(0),
		Keepers:     signer.Keys{},
		Tokens:      make(map[string]*big.Int),
		NonceTable:  make(map[uint64][]uint64),
		ApproveList: TxTypes{TypeTransfer, TypeTransfer},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid approveList, duplicate TxType TypeTransfer")

	acc = &Account{
		Type:       NativeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Stake:      &StakeConfig{},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid stake on NativeAccount")

	acc = &Account{
		Type:       TokenAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Stake:      &StakeConfig{},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid stake on TokenAccount")

	acc = &Account{
		Type:       TokenAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid maxTotalSupply")

	acc = &Account{
		Type:           TokenAccount,
		Balance:        big.NewInt(0),
		Keepers:        signer.Keys{},
		Tokens:         make(map[string]*big.Int),
		NonceTable:     make(map[uint64][]uint64),
		MaxTotalSupply: big.NewInt(-1),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid maxTotalSupply")

	acc = &Account{
		Type:           StakeAccount,
		Balance:        big.NewInt(0),
		Keepers:        signer.Keys{},
		Tokens:         make(map[string]*big.Int),
		NonceTable:     make(map[uint64][]uint64),
		MaxTotalSupply: big.NewInt(0),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid maxTotalSupply, should be nil")

	acc = &Account{
		Type:       StakeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Stake:      &StakeConfig{},
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid withdrawFee, should be in [1, 200_000]")

	acc = &Account{
		Type:       StakeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.ErrorContains(acc.SyntacticVerify(), "invalid stake on StakeAccount")

	acc = &Account{
		Type:       StakeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Stake: &StakeConfig{
			WithdrawFee: 1,
			MinAmount:   new(big.Int).SetUint64(100),
			MaxAmount:   new(big.Int).SetUint64(100),
		},
		Lending: &LendingConfig{
			DailyInterest:   10,
			OverdueInterest: 1,
			MinAmount:       new(big.Int).SetUint64(100),
			MaxAmount:       new(big.Int).SetUint64(100),
		},
	}
	assert.NoError(acc.SyntacticVerify())

	acc = &Account{
		Type:       StakeAccount,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
		Stake: &StakeConfig{
			WithdrawFee: 1,
			MinAmount:   new(big.Int).SetUint64(100),
			MaxAmount:   new(big.Int).SetUint64(100),
		},
		Lending: &LendingConfig{
			DailyInterest:   10,
			OverdueInterest: 1,
			MinAmount:       new(big.Int).SetUint64(100),
			MaxAmount:       new(big.Int).SetUint64(100),
		},
	}
	assert.NoError(acc.SyntacticVerify())
	cbordata, err := acc.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(acc)
	require.NoError(t, err)

	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"Stake","nonce":0,"balance":0,"threshold":0,"keepers":[],"tokens":{},"nonceTable":{},"stake":{"token":"","type":0,"lockTime":0,"withdrawFee":1,"minAmount":100,"maxAmount":100},"lending":{"token":"","dailyInterest":10,"overdueInterest":1,"minAmount":100,"maxAmount":100},"height":0,"timestamp":0,"address":"0x0000000000000000000000000000000000000000"}`, string(jsondata))

	acc2 := &Account{}
	assert.NoError(acc2.Unmarshal(cbordata))
	assert.NoError(acc2.SyntacticVerify())

	cbordata2, err := acc2.Marshal()
	require.NoError(t, err)
	jsondata2, err := json.Marshal(acc2)
	require.NoError(t, err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
