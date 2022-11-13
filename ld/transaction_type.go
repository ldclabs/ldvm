// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import "fmt"

const (
	// The "test" transaction tests that a value of data at the target location
	// is equal to a specified value.
	// It should be in a batch txn.
	TypeTest TxType = iota

	// Transfer
	TypeEth              // Sends token to a address in ETH transaction
	TypeTransfer         // Sends token to a address
	TypeTransferPay      // Sends token to the address who request payment
	TypeTransferCash     // Transfer token to sender, like cashing a check.
	TypeTransferMultiple // Sends token to multiple addresses.
	TypeExchange         // Exchanges tokens
)

const (
	// Punishs transaction can be issued by genesisAccount
	// we can only punish illegal data
	TypePunish TxType = 16 + iota

	// Model
	TypeCreateModel     // Creates a data model
	TypeUpdateModelInfo // Updates data model's info

	// Data
	TypeCreateData           // Creates a data from the model
	TypeUpdateData           // Updates the data's data
	TypeUpgradeData          // Updates the data's model and data
	TypeUpdateDataInfo       // Updates data's info, such as keepers, threshold, approvers, sigClaims, etc.
	TypeUpdateDataInfoByAuth // Updates data's info by authorization
	TypeDeleteData           // Deletes the data
)

const (
	// Account
	TypeUpdateNonceTable  TxType = 32 + iota // Add or update nonce with expire time
	TypeUpdateAccountInfo                    // Updates account's Keepers and Threshold
	TypeCreateToken                          // Creates a token account
	TypeDestroyToken                         // Destroy a token account
	TypeCreateStake                          // Creates a stake account
	TypeResetStake                           // Reset a stake account
	TypeDestroyStake                         // Destroy a stake account
	TypeTakeStake                            // take a stake in
	TypeWithdrawStake                        // Withdraw stake
	TypeUpdateStakeApprover
	TypeOpenLending
	TypeCloseLending
	TypeBorrow
	TypeRepay
)

// TxTypes set
var TransferTxTypes = TxTypes{
	TypeEth,
	TypeTransfer,
	TypeTransferPay,
	TypeTransferCash,
	TypeTransferMultiple,
	TypeExchange,
}

var ModelTxTypes = TxTypes{
	TypeUpdateModelInfo,
}

var DataTxTypes = TxTypes{
	TypeUpdateData,
	TypeUpgradeData,
	TypeUpdateDataInfo,
	TypeUpdateDataInfoByAuth,
	TypeDeleteData,
}

var AccountTxTypes = TxTypes{
	TypeUpdateNonceTable,
	TypeUpdateAccountInfo,
	TypeCreateToken,
	TypeDestroyToken,
	TypeCreateStake,
	TypeResetStake,
	TypeDestroyStake,
	TypeTakeStake,
	TypeWithdrawStake,
	TypeUpdateStakeApprover,
	TypeOpenLending,
	TypeCloseLending,
	TypeBorrow,
	TypeRepay,
}

var AllTxTypes = TxTypes{
	TypeTest,
	TypePunish,
	TypeCreateModel,
	TypeCreateData,
}.Union(
	TransferTxTypes,
	ModelTxTypes,
	DataTxTypes,
	AccountTxTypes,
)

var TokenFromTxTypes = TxTypes{
	TypeEth,
	TypeTransfer,
	TypeUpdateAccountInfo,
	TypeUpdateNonceTable,
	TypeDestroyToken,
	TypeOpenLending,
	TypeCloseLending,
}

var TokenToTxTypes = TxTypes{
	TypeTest,
	TypeEth,
	TypeTransfer,
	TypeExchange,
	TypeCreateToken,
	TypeBorrow,
	TypeRepay,
}

var StakeFromTxTypes0 = TxTypes{
	TypeUpdateAccountInfo,
	TypeUpdateNonceTable,
	TypeResetStake,
	TypeDestroyStake,
}

var StakeFromTxTypes1 = TxTypes{
	TypeTakeStake,
	TypeWithdrawStake,
	TypeUpdateStakeApprover,
	TypeOpenLending,
	TypeCloseLending,
}.Union(StakeFromTxTypes0)

var StakeFromTxTypes2 = TxTypes{
	TypeEth,
	TypeTransfer,
	TypeTransferMultiple,
}.Union(StakeFromTxTypes1)

var StakeToTxTypes = TxTypes{
	TypeTest,
	TypeEth,
	TypeTransfer,
	TypeCreateStake,
	TypeTakeStake,
	TypeWithdrawStake,
	TypeUpdateStakeApprover,
	TypeBorrow,
	TypeRepay,
}

// TxType is an uint16 representing the type of the tx.
// to avoid encode/decode TxTypes as []uint8, aka []byte.
type TxType uint16

func (t TxType) Gas() uint64 {
	switch t {
	case TypeTest:
		return 0

	case TypeEth, TypeTransfer, TypeTransferPay, TypeTransferCash, TypeTransferMultiple, TypeExchange:
		return 42

	case TypeUpdateNonceTable, TypeUpdateAccountInfo, TypeUpdateData, TypeUpdateDataInfo:
		return 42

	case TypePunish, TypeCreateData, TypeUpgradeData, TypeUpdateDataInfoByAuth, TypeDeleteData:
		return 200

	case TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover:
		return 200

	case TypeBorrow, TypeRepay:
		return 500

	case TypeCreateModel, TypeUpdateModelInfo:
		return 500

	case TypeCreateToken, TypeDestroyToken, TypeCreateStake, TypeResetStake, TypeDestroyStake:
		return 1000

	case TypeOpenLending, TypeCloseLending:
		return 1000

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
	case TypeTransferMultiple:
		return "TypeTransferMultiple"
	case TypeExchange:
		return "TypeExchange"
	case TypeUpdateNonceTable:
		return "TypeUpdateNonceTable"
	case TypeUpdateAccountInfo:
		return "TypeUpdateAccountInfo"
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
	case TypeUpdateModelInfo:
		return "TypeUpdateModelInfo"
	case TypeCreateData:
		return "TypeCreateData"
	case TypeUpdateData:
		return "TypeUpdateData"
	case TypeUpgradeData:
		return "TypeUpgradeData"
	case TypeUpdateDataInfo:
		return "TypeUpdateDataInfo"
	case TypeUpdateDataInfoByAuth:
		return "TypeUpdateDataInfoByAuth"
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
