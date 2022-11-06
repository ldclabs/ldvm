// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util/signer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLedger(t *testing.T) {
	assert := assert.New(t)

	var al *AccountLedger
	assert.ErrorContains(al.SyntacticVerify(), "nil pointer")

	al = &AccountLedger{}
	assert.NoError(al.SyntacticVerify())

	al = &AccountLedger{
		Lending: map[string]*LendingEntry{
			constants.GenesisAccount.AsKey(): nil,
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on LendingEntry")

	al = &AccountLedger{
		Lending: map[string]*LendingEntry{
			constants.GenesisAccount.AsKey(): {Amount: nil},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on LendingEntry")

	al = &AccountLedger{
		Lending: map[string]*LendingEntry{
			constants.GenesisAccount.AsKey(): {Amount: big.NewInt(-1)},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on LendingEntry")

	al = &AccountLedger{
		Lending: map[string]*LendingEntry{
			constants.GenesisAccount.AsKey(): {Amount: big.NewInt(0)},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on LendingEntry")

	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): nil,
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on StakeEntry")

	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {Amount: nil},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on StakeEntry")

	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {Amount: big.NewInt(-1)},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on StakeEntry")

	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {Amount: big.NewInt(0)},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid amount on StakeEntry")

	key := signer.Key(constants.LDCAccount[:])
	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {
				Amount: big.NewInt(1), Approver: &key},
		},
	}
	assert.ErrorContains(al.SyntacticVerify(), "invalid approver on StakeEntry")

	key = signer.Key(constants.GenesisAccount[:])
	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {
				Amount: big.NewInt(0), Approver: &key},
		},
	}
	assert.NoError(al.SyntacticVerify())

	al = &AccountLedger{
		Stake: map[string]*StakeEntry{
			constants.GenesisAccount.AsKey(): {
				LockTime: 999,
				Amount:   new(big.Int).SetUint64(100),
				Approver: &key,
			},
		},
		Lending: map[string]*LendingEntry{
			constants.GenesisAccount.AsKey(): {
				Amount:   new(big.Int).SetUint64(100),
				UpdateAt: 888,
			},
		},
	}
	assert.NoError(al.SyntacticVerify())
	cbordata, err := al.Marshal()
	require.NoError(t, err)

	al2 := &AccountLedger{}
	assert.NoError(al2.Unmarshal(cbordata))
	assert.NoError(al2.SyntacticVerify())
	assert.Equal(cbordata, al2.Bytes())
}
