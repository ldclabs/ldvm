// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import "fmt"

const (
	// The "test" transaction tests that a value of data at the target location
	// is equal to a specified value. test transaction will not write to the block.
	// It should be in a batch transactions.
	TypeTest TxType = iota

	// Transfer
	TypeEth          // send given amount of NanoLDC to a address in ETH transaction
	TypeTransfer     // send given amount of NanoLDC to a address
	TypeTransferPay  // send given amount of NanoLDC to the address who request payment
	TypeTransferCash // cash given amount of NanoLDC to sender, like cashing a check.
	TypeExchange     // exchange tokens
)

const (
	// punish transaction can be issued by genesisAccount
	// we can only punish illegal data
	TypePunish TxType = 16 + iota

	// Model
	TypeCreateModel        // create a data model
	TypeUpdateModelKeepers // update data model's Keepers and Threshold

	// Data
	TypeCreateData              // create a data from the model
	TypeUpdateData              // update the data's Data
	TypeUpdateDataKeepers       // update data's Keepers and Threshold
	TypeUpdateDataKeepersByAuth // update data's Keepers and Threshold by authorization
	TypeDeleteData              // delete the data
)

const (
	// Account
	TypeAddNonceTable        TxType = 32 + iota // add more nonce with expire time to account
	TypeUpdateAccountKeepers                    // update account's Keepers and Threshold
	TypeCreateToken                             // create a token account
	TypeDestroyToken                            // destroy a token account
	TypeCreateStake                             // create a stake account
	TypeResetStake                              // reset a stake account
	TypeDestroyStake                            // destroy a stake account
	TypeTakeStake                               // take a stake in
	TypeWithdrawStake                           // withdraw stake
	TypeUpdateStakeApprover
	TypeOpenLending
	TypeCloseLending
	TypeBorrow
	TypeRepay
)

// TxTypes set
var TransferTxTypes = TxTypes{TypeEth, TypeTransfer, TypeTransferPay, TypeTransferCash, TypeExchange}
var ModelTxTypes = TxTypes{TypeUpdateModelKeepers}
var DataTxTypes = TxTypes{TypeUpdateData, TypeUpdateDataKeepers, TypeUpdateDataKeepersByAuth, TypeDeleteData}
var AccountTxTypes = TxTypes{TypeAddNonceTable, TypeUpdateAccountKeepers, TypeCreateToken,
	TypeDestroyToken, TypeCreateStake, TypeResetStake, TypeDestroyStake, TypeTakeStake,
	TypeWithdrawStake, TypeUpdateStakeApprover, TypeOpenLending, TypeCloseLending, TypeBorrow, TypeRepay}
var AllTxTypes = TxTypes{TypeTest, TypePunish, TypeCreateModel, TypeCreateData}.Union(
	TransferTxTypes, ModelTxTypes, DataTxTypes, AccountTxTypes)

var TokenFromTxTypes = TxTypes{TypeEth, TypeTransfer, TypeUpdateAccountKeepers, TypeAddNonceTable, TypeDestroyToken, TypeOpenLending, TypeCloseLending}
var TokenToTxTypes = TxTypes{TypeTest, TypeEth, TypeTransfer, TypeExchange, TypeCreateToken, TypeBorrow, TypeRepay}
var StakeFromTxTypes0 = TxTypes{TypeUpdateAccountKeepers, TypeAddNonceTable, TypeResetStake}
var StakeFromTxTypes1 = TxTypes{TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover, TypeOpenLending, TypeCloseLending}.Union(StakeFromTxTypes0)
var StakeFromTxTypes2 = TxTypes{TypeEth, TypeTransfer}.Union(StakeFromTxTypes1)
var StakeToTxTypes = TxTypes{TypeTest, TypeEth, TypeTransfer, TypeCreateStake, TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover, TypeBorrow, TypeRepay}

// TxType is an uint16 representing the type of the tx.
// to avoid encode/decode TxTypes as []uint8, aka []byte.
type TxType uint16

func (t TxType) Gas() uint64 {
	switch t {
	case TypeTest:
		return 0
	case TypeEth, TypeTransfer, TypeTransferPay, TypeTransferCash,
		TypeExchange, TypeAddNonceTable:
		return 42
	case TypeUpdateAccountKeepers, TypeCreateToken,
		TypeDestroyToken, TypeCreateStake, TypeResetStake, TypeDestroyStake:
		return 1000
	case TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover:
		return 500
	case TypeOpenLending, TypeCloseLending:
		return 1000
	case TypeBorrow, TypeRepay:
		return 500
	case TypePunish:
		return 42
	case TypeCreateModel, TypeUpdateModelKeepers:
		return 500
	case TypeCreateData, TypeUpdateData, TypeUpdateDataKeepers:
		return 100
	case TypeUpdateDataKeepersByAuth, TypeDeleteData:
		return 200
	default:
		return 10000
	}
}

func (t TxType) String() string {
	switch t {
	case TypeTest:
		return "TypeTest"
	case TypePunish:
		return "TypePunish"
	case TypeEth:
		return "TypeEth"
	case TypeTransfer:
		return "TypeTransfer"
	case TypeTransferPay:
		return "TypeTransferPay"
	case TypeTransferCash:
		return "TypeTransferCash"
	case TypeExchange:
		return "TypeExchange"
	case TypeAddNonceTable:
		return "TypeAddNonceTable"
	case TypeUpdateAccountKeepers:
		return "TypeUpdateAccountKeepers"
	case TypeCreateToken:
		return "TypeCreateToken"
	case TypeDestroyToken:
		return "TypeDestroyToken"
	case TypeCreateStake:
		return "TypeCreateStake"
	case TypeResetStake:
		return "TypeResetStake"
	case TypeDestroyStake:
		return "TypeDestroyStake"
	case TypeTakeStake:
		return "TypeTakeStake"
	case TypeWithdrawStake:
		return "TypeWithdrawStake"
	case TypeUpdateStakeApprover:
		return "TypeUpdateStakeApprover"
	case TypeOpenLending:
		return "TypeOpenLending"
	case TypeCloseLending:
		return "TypeCloseLending"
	case TypeBorrow:
		return "TypeBorrow"
	case TypeRepay:
		return "TypeRepay"
	case TypeCreateModel:
		return "TypeCreateModel"
	case TypeUpdateModelKeepers:
		return "TypeUpdateModelKeepers"
	case TypeCreateData:
		return "TypeCreateData"
	case TypeUpdateData:
		return "TypeUpdateData"
	case TypeUpdateDataKeepers:
		return "TypeUpdateDataKeepers"
	case TypeUpdateDataKeepersByAuth:
		return "TypeUpdateDataKeepersByAuth"
	case TypeDeleteData:
		return "TypeDeleteData"
	default:
		return fmt.Sprintf("TypeUnknown(%d)", t)
	}
}

func (t TxType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

type TxTypes []TxType

func (ts TxTypes) Has(ty TxType) bool {
	for _, t := range ts {
		if t == ty {
			return true
		}
	}
	return false
}

func (ts TxTypes) Union(tss ...TxTypes) TxTypes {
	for _, vv := range tss {
		ts = append(ts, vv...)
	}
	return ts
}

func (ts TxTypes) CheckDuplicate() error {
	set := make(map[TxType]struct{}, len(ts))
	for _, v := range ts {
		if _, ok := set[v]; ok {
			return fmt.Errorf("duplicate TxType %s", v)
		}
		set[v] = struct{}{}
	}
	return nil
}
