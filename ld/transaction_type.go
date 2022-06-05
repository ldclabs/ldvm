// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

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
var ModelTxTypes = TxTypes{TypeCreateModel, TypeUpdateModelKeepers}
var DataTxTypes = TxTypes{TypeCreateData, TypeUpdateData, TypeUpdateDataKeepers, TypeUpdateDataKeepersByAuth, TypeDeleteData}
var AccountTxTypes = TxTypes{TypeAddNonceTable, TypeUpdateAccountKeepers, TypeCreateToken,
	TypeDestroyToken, TypeCreateStake, TypeResetStake, TypeDestroyStake, TypeTakeStake,
	TypeWithdrawStake, TypeUpdateStakeApprover, TypeOpenLending, TypeCloseLending, TypeBorrow, TypeRepay}
var AllTxTypes = TxTypes{TypeTest, TypePunish}.Union(
	TransferTxTypes, ModelTxTypes, DataTxTypes, AccountTxTypes)

var TokenFromTxTypes = TxTypes{TypeEth, TypeTransfer, TypeUpdateAccountKeepers, TypeAddNonceTable, TypeDestroyToken, TypeOpenLending, TypeCloseLending}
var TokenToTxTypes = TxTypes{TypeTest, TypeEth, TypeTransfer, TypeExchange, TypeCreateToken}
var StakeFromTxTypes0 = TxTypes{TypeUpdateAccountKeepers, TypeAddNonceTable, TypeResetStake, TypeBorrow, TypeRepay}
var StakeFromTxTypes1 = TxTypes{TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover, TypeOpenLending, TypeCloseLending}.Union(StakeFromTxTypes0)
var StakeFromTxTypes2 = TxTypes{TypeEth, TypeTransfer}.Union(StakeFromTxTypes1)
var StakeToTxTypes = TxTypes{TypeTest, TypeEth, TypeTransfer, TypeCreateStake, TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover}

// TxType is an uint8 representing the type of the tx
type TxType uint8

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
		return "TestTx"
	case TypePunish:
		return "PunishTx"
	case TypeEth:
		return "EthTx"
	case TypeTransfer:
		return "TransferTx"
	case TypeTransferPay:
		return "TransferPayTx"
	case TypeTransferCash:
		return "TransferCashTx"
	case TypeExchange:
		return "ExchangeTx"
	case TypeAddNonceTable:
		return "TypeAddNonceTable"
	case TypeUpdateAccountKeepers:
		return "UpdateAccountKeepersTx"
	case TypeCreateToken:
		return "CreateTokenTx"
	case TypeDestroyToken:
		return "DestroyTokenTx"
	case TypeCreateStake:
		return "CreateStakeTx"
	case TypeResetStake:
		return "ResetStakeTx"
	case TypeDestroyStake:
		return "DestroyStakeTx"
	case TypeTakeStake:
		return "TakeStakeTx"
	case TypeWithdrawStake:
		return "WithdrawStakeTx"
	case TypeUpdateStakeApprover:
		return "TypeUpdateStakeApprover"
	case TypeOpenLending:
		return "OpenLendingTx"
	case TypeCloseLending:
		return "CloseLendingTx"
	case TypeBorrow:
		return "BorrowTx"
	case TypeRepay:
		return "RepayTx"
	case TypeCreateModel:
		return "CreateModelTx"
	case TypeUpdateModelKeepers:
		return "UpdateModelKeepersTx"
	case TypeCreateData:
		return "CreateDataTx"
	case TypeUpdateData:
		return "UpdateDataTx"
	case TypeUpdateDataKeepers:
		return "UpdateDataKeepersTx"
	case TypeUpdateDataKeepersByAuth:
		return "UpdateDataKeepersByAuthTx"
	case TypeDeleteData:
		return "DeleteDataTx"
	default:
		return "UnknownTx"
	}
}

// func (t TxType) MarshalJSON() ([]byte, error) {
// 	return []byte("\"" + t.String() + "\""), nil
// }

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
