// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

var (
	_ LDObject = (*LendingConfig)(nil)
	_ LDObject = (*StakeConfig)(nil)
)

// AccountType is an uint8 representing the type of account
type AccountType uint8

const (
	NativeAccount AccountType = iota
	TokenAccount              // The first 10 bytes of account address must be 0
	StakeAccount              // The first byte of account address must be $
)

func AccountTypeString(t AccountType) string {
	switch t {
	case NativeAccount:
		return "NativeAccount"
	case TokenAccount:
		return "TokenAccount"
	case StakeAccount:
		return "StakeAccount"
	default:
		return "TypeUnknown"
	}
}

type Account struct {
	Type AccountType
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int
	// M of N threshold signatures, aka MultiSig: threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint8
	// keepers who can use this account, no more than 255
	// the account id must be one of them.
	Keepers []ids.ShortID

	MaxTotalSupply *big.Int            // only used with TokenAccount
	NonceTable     map[uint64][]uint64 // map[expire][]nonce
	Tokens         map[ids.ShortID]*big.Int

	Stake         *StakeConfig
	StakeLedger   Ledger
	Lending       *LendingConfig
	LendingLedger Ledger

	// external assignment
	Height    uint64 // block's timestamp
	Timestamp uint64 // block's timestamp
	raw       []byte
	ID        ids.ShortID
}

type jsonAccount struct {
	Type           string                  `json:"type"`
	Address        string                  `json:"address"`
	Nonce          uint64                  `json:"nonce"`
	Balance        *big.Int                `json:"balance"`
	Threshold      uint8                   `json:"threshold"`
	Keepers        []string                `json:"keepers"`
	MaxTotalSupply *big.Int                `json:"maxTotalSupply,omitempty"`
	NonceTable     map[string][]uint64     `json:"nonceTable"`
	Tokens         map[string]*big.Int     `json:"tokens"`
	Stake          *StakeConfig            `json:"stake,omitempty"`
	StakeLedger    map[string]*LedgerEntry `json:"stakeLedger,omitempty"`
	Lending        *LendingConfig          `json:"lending,omitempty"`
	LendingLedger  map[string]*LedgerEntry `json:"lendingLedger,omitempty"`
}

type Ledger map[ids.ShortID]*LedgerEntry

func LedgerFromBytesList(list [][]byte) (Ledger, error) {
	ledger := make(map[ids.ShortID]*LedgerEntry, len(list))
	for _, data := range list {
		list2, err := ReadBytesList(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("unmarshal error: %v", err)
		}
		if len(list2) != 4 || len(list2[0]) != 20 {
			return nil, fmt.Errorf("unmarshal error: invalid LedgerEntry data")
		}
		le, err := LedgerEntryFromBytesList(list2[1:])
		if err != nil {
			return nil, fmt.Errorf("unmarshal error: %v", err)
		}
		id := ids.ShortID{}
		copy(id[:], list[0])
		ledger[id] = le
	}
	return ledger, nil
}

func (l Ledger) ToBytesList(buf *bytes.Buffer) ([][]byte, error) {
	if l == nil {
		return [][]byte{}, nil
	}
	var err error
	list := make([][]byte, 0, len(l))
	if buf == nil {
		buf = new(bytes.Buffer)
	}
	for k, v := range l {
		if err = WriteBytesList(buf, k[:], v.ToBytesList()...); err != nil {
			return nil, err
		}

		list = append(list, buf.Bytes())
		buf.Reset()
	}
	return list, nil
}

type LedgerEntry struct {
	Amount   *big.Int `json:"amount"`
	UpdateAt uint64   `json:"updateAt"`
	DueTime  uint64   `json:"dueTime"`
}

func (e *LedgerEntry) ToBytesList() [][]byte {
	return [][]byte{FromUint(e.Amount), FromUint64(e.UpdateAt), FromUint64(e.DueTime)}
}

func LedgerEntryFromBytesList(list [][]byte) (*LedgerEntry, error) {
	if len(list) != 3 {
		return nil, fmt.Errorf("invalid LedgerEntry bytes list")
	}
	e := &LedgerEntry{}
	e.Amount = new(big.Int).SetBytes(list[0])
	u := Uint64(list[1])
	if !u.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	e.UpdateAt = u.Value()
	u = Uint64(list[2])
	if !u.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	e.DueTime = u.Value()
	return e, nil
}

type StakeConfig struct {
	// 0: account keepers can not use stake token
	// 1: account keepers can take a stake in other stake account
	// 2: in addition to 1, account keepers can transfer stake token to other account
	// 3: in addition to 2, account keepers can liquidate shareholder (use with lending)
	Type        uint8    `json:"type"`
	Token       string   `json:"token"`
	LockTime    uint64   `json:"lockTime"`
	WithdrawFee uint64   `json:"withdrawFee"` // 1_000_000 == 100%, should be in [1, 200_000]
	MinAmount   *big.Int `json:"minAmount"`
	MaxAmount   *big.Int `json:"maxAmount"`

	TokenID ids.ShortID `json:"-"`
}

// SyntacticVerify verifies that a *StakeConfig is well-formed.
func (c *StakeConfig) SyntacticVerify() error {
	if c == nil {
		return fmt.Errorf("invalid StakeConfig")
	}
	if c.Type > 3 {
		return fmt.Errorf("invalid StakeConfig type")
	}
	token, err := util.NewSymbol(c.Token)
	if err != nil {
		return err
	}
	c.TokenID = ids.ShortID(token)

	if c.WithdrawFee < 1 || c.WithdrawFee > 200_000 {
		return fmt.Errorf("invalid WithdrawFee, should be in [1, 200_000]")
	}

	if c.MinAmount == nil || c.MinAmount.Sign() < 1 {
		return fmt.Errorf("invalid MinAmount")
	}
	if c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0 {
		return fmt.Errorf("invalid MaxAmount")
	}
	return nil
}

func (c *StakeConfig) Unmarshal(data []byte) error {
	list, err := ReadBytesList(bytes.NewReader(data))
	if err != nil {
		return err
	}
	lc, err := StakeConfigFromBytesList(list)
	if err != nil {
		return err
	}
	*c = *lc
	return nil
}

func (c *StakeConfig) Marshal() ([]byte, error) {
	list := c.ToBytesList()
	buf := new(bytes.Buffer)
	if err := WriteBytesList(buf, list[0], list[1:]...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *StakeConfig) MarshalJSON() ([]byte, error) {
	if c == nil {
		return util.Null, nil
	}
	return json.Marshal(c)
}

func (c *StakeConfig) ToBytesList() [][]byte {
	return [][]byte{
		FromUint8(c.Type),
		[]byte(c.Token),
		FromUint64(c.LockTime),
		FromUint64(c.WithdrawFee),
		FromUint(c.MinAmount),
		FromUint(c.MaxAmount),
	}
}

func StakeConfigFromBytesList(list [][]byte) (*StakeConfig, error) {
	if len(list) != 6 {
		return nil, fmt.Errorf("invalid StakeConfig bytes list")
	}
	c := &StakeConfig{}
	u := Uint8(list[0])
	if !u.Valid() {
		return nil, fmt.Errorf("invalid uint8 bytes")
	}
	c.Type = u.Value()
	c.Token = string(list[1])
	u2 := Uint64(list[2])
	if !u2.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	c.LockTime = u2.Value()
	u2 = Uint64(list[3])
	if !u2.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	c.WithdrawFee = u2.Value()
	c.MinAmount = new(big.Int).SetBytes(list[4])
	c.MaxAmount = new(big.Int).SetBytes(list[5])
	return c, nil
}

type LendingConfig struct {
	Token           string   `json:"token"`
	DailyInterest   uint64   `json:"dailyInterest"`   // 1_000_000 == 100%, should be in [1, 10_000]
	OverdueInterest uint64   `json:"overdueInterest"` // 1_000_000 == 100%, should be in [1, 10_000]
	MinAmount       *big.Int `json:"minAmount"`
	MaxAmount       *big.Int `json:"maxAmount"`

	TokenID ids.ShortID `json:"-"`
}

// SyntacticVerify verifies that a *LendingConfig is well-formed.
func (c *LendingConfig) SyntacticVerify() error {
	if c == nil {
		return fmt.Errorf("invalid LendingConfig")
	}
	token, err := util.NewSymbol(c.Token)
	if err != nil {
		return err
	}
	c.TokenID = ids.ShortID(token)

	if c.DailyInterest < 1 || c.DailyInterest > 10_000 {
		return fmt.Errorf("invalid DailyInterest, should be in [1, 10_000]")
	}
	if c.OverdueInterest < 1 || c.OverdueInterest > 10_000 {
		return fmt.Errorf("invalid OverdueInterest, should be in [1, 10_000]")
	}

	if c.MinAmount == nil || c.MinAmount.Sign() < 1 {
		return fmt.Errorf("invalid MinAmount")
	}
	if c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0 {
		return fmt.Errorf("invalid MaxAmount")
	}
	return nil
}

func (c *LendingConfig) Unmarshal(data []byte) error {
	list, err := ReadBytesList(bytes.NewReader(data))
	if err != nil {
		return err
	}
	lc, err := LendingConfigFromBytesList(list)
	if err != nil {
		return err
	}
	*c = *lc
	return nil
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	list := c.ToBytesList()
	buf := new(bytes.Buffer)
	if err := WriteBytesList(buf, list[0], list[1:]...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *LendingConfig) MarshalJSON() ([]byte, error) {
	if c == nil {
		return util.Null, nil
	}
	return json.Marshal(c)
}

func (c *LendingConfig) ToBytesList() [][]byte {
	return [][]byte{
		[]byte(c.Token),
		FromUint64(c.DailyInterest),
		FromUint64(c.OverdueInterest),
		FromUint(c.MinAmount),
		FromUint(c.MaxAmount),
	}
}

func LendingConfigFromBytesList(list [][]byte) (*LendingConfig, error) {
	if len(list) != 5 {
		return nil, fmt.Errorf("invalid LendingConfig bytes list")
	}
	c := &LendingConfig{}
	c.Token = string(list[0])
	u := Uint64(list[1])
	if !u.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	c.DailyInterest = u.Value()
	u = Uint64(list[2])
	if !u.Valid() {
		return nil, fmt.Errorf("invalid uint64 bytes")
	}
	c.OverdueInterest = u.Value()
	c.MinAmount = new(big.Int).SetBytes(list[3])
	c.MaxAmount = new(big.Int).SetBytes(list[4])
	return c, nil
}

func (a *Account) MarshalJSON() ([]byte, error) {
	if a == nil {
		return util.Null, nil
	}
	v := &jsonAccount{
		Type:           AccountTypeString(a.Type),
		Address:        util.EthID(a.ID).String(),
		Nonce:          a.Nonce,
		Balance:        a.Balance,
		Threshold:      a.Threshold,
		Keepers:        make([]string, len(a.Keepers)),
		MaxTotalSupply: a.MaxTotalSupply,
		NonceTable:     make(map[string][]uint64, len(a.NonceTable)),
		Tokens:         make(map[string]*big.Int, len(a.Tokens)),
		Stake:          a.Stake,
		Lending:        a.Lending,
	}
	for i := range a.Keepers {
		v.Keepers[i] = util.EthID(a.Keepers[i]).String()
	}
	for k, arr := range a.NonceTable {
		v.NonceTable[strconv.FormatUint(k, 10)] = arr
	}
	for k := range a.Tokens {
		str := util.TokenSymbol(k).String()
		if str == "" {
			str = util.EthID(k).String()
		}
		v.Tokens[str] = a.Tokens[k]
	}
	if a.Stake != nil {
		v.StakeLedger = make(map[string]*LedgerEntry, len(a.StakeLedger))
		for k := range a.StakeLedger {
			v.StakeLedger[util.EthID(k).String()] = a.StakeLedger[k]
		}
	}
	if a.Lending != nil {
		v.LendingLedger = make(map[string]*LedgerEntry, len(a.LendingLedger))
		for k := range a.LendingLedger {
			v.LendingLedger[util.EthID(k).String()] = a.LendingLedger[k]
		}
	}
	return json.Marshal(v)
}

func (a *Account) Copy() *Account {
	x := new(Account)
	*x = *a
	x.Balance = new(big.Int).Set(a.Balance)
	x.Keepers = make([]ids.ShortID, len(a.Keepers))
	copy(x.Keepers, a.Keepers)
	x.NonceTable = make(map[uint64][]uint64, len(a.NonceTable))
	for k, v := range a.NonceTable {
		x.NonceTable[k] = make([]uint64, len(v))
		for i := range v {
			x.NonceTable[k][i] = v[i]
		}
	}
	x.Tokens = make(map[ids.ShortID]*big.Int, len(a.Tokens))
	for k, v := range a.Tokens {
		x.Tokens[k] = new(big.Int).Set(v)
	}
	if a.MaxTotalSupply != nil {
		x.MaxTotalSupply = new(big.Int).Set(a.MaxTotalSupply)
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	if a == nil {
		return fmt.Errorf("invalid Account")
	}

	if a.Balance == nil || a.Balance.Sign() < 0 {
		return fmt.Errorf("invalid balance")
	}
	if len(a.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(a.Threshold) > len(a.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	if a.NonceTable == nil {
		return fmt.Errorf("invalid nonceTable")
	}
	if a.Tokens == nil {
		return fmt.Errorf("invalid tokens")
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake != nil || len(a.StakeLedger) > 0 {
			return fmt.Errorf("invalid stake on NativeAccount")
		}
	case TokenAccount:
		if a.Stake != nil || len(a.StakeLedger) > 0 {
			return fmt.Errorf("invalid stake on TokenAccount")
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return fmt.Errorf("invalid maxTotalSupply")
		}
	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake == nil || a.StakeLedger == nil {
			return fmt.Errorf("invalid stake on StakeAccount")
		}
	default:
		return fmt.Errorf("invalid type")
	}

	if _, err := a.Marshal(); err != nil {
		return fmt.Errorf("Account marshal error: %v", err)
	}
	return nil
}

func (a *Account) Equal(o *Account) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(a.raw) > 0 {
		return bytes.Equal(o.raw, a.raw)
	}
	if o.Type != a.Type {
		return false
	}
	if o.Nonce != a.Nonce {
		return false
	}
	if o.Balance == nil || a.Balance == nil || o.Balance.Cmp(a.Balance) != 0 {
		return false
	}
	if o.MaxTotalSupply == nil || a.MaxTotalSupply == nil {
		if o.MaxTotalSupply != a.MaxTotalSupply {
			return false
		}
	} else if o.MaxTotalSupply.Cmp(a.MaxTotalSupply) != 0 {
		return false
	}
	if o.Threshold != a.Threshold {
		return false
	}
	if len(o.Keepers) != len(a.Keepers) {
		return false
	}
	for i := range a.Keepers {
		if o.Keepers[i] != a.Keepers[i] {
			return false
		}
	}
	if len(o.NonceTable) != len(a.NonceTable) {
		return false
	}
	for k, v := range a.NonceTable {
		if len(o.NonceTable[k]) != len(v) {
			return false
		}
		for i := range v {
			if o.NonceTable[k][i] != v[i] {
				return false
			}
		}
	}
	if len(o.Tokens) != len(a.Tokens) {
		return false
	}
	for id := range a.Tokens {
		if o.Tokens[id] == nil || a.Tokens[id] == nil || o.Tokens[id].Cmp(a.Tokens[id]) != 0 {
			return false
		}
	}
	return true
}

func (a *Account) Bytes() []byte {
	if len(a.raw) == 0 {
		if _, err := a.Marshal(); err != nil {
			panic(err)
		}
	}

	return a.raw
}

func (a *Account) Unmarshal(data []byte) error {
	p, err := accountLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindAccount); ok {
		if !v.Type.Valid() || !v.Threshold.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}
		if !v.Nonce.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}

		a.Type = AccountType(v.Type.Value())
		a.Nonce = v.Nonce.Value()
		a.Balance = v.Balance.Value()
		a.Threshold = v.Threshold.Value()
		a.MaxTotalSupply = v.MaxTotalSupply.Value()
		a.NonceTable = make(map[uint64][]uint64, len(a.NonceTable))

		if a.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}

		for _, data := range v.NonceTable {
			uu, err := ReadUint64s(data)
			if err != nil {
				return err
			}
			if len(uu) < 1 {
				return fmt.Errorf("unmarshal NonceTable error")
			}
			if exp := uu[0]; exp >= a.Timestamp && len(uu) > 1 {
				a.NonceTable[exp] = uu[1:]
			}
		}

		a.Tokens = make(map[ids.ShortID]*big.Int, len(v.Tokens))
		for _, data := range v.Tokens {
			list, err := ReadBytesList(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("unmarshal error: %v", err)
			}
			if len(list) != 2 || len(list[0]) != 20 {
				return fmt.Errorf("unmarshal error: invalid Tokens data")
			}

			id := ids.ShortID{}
			copy(id[:], list[0])
			a.Tokens[id] = new(big.Int).SetBytes(list[1])
		}

		if v.Stake != nil {
			a.Stake, err = StakeConfigFromBytesList(*v.Stake)
			if err != nil {
				return fmt.Errorf("unmarshal error: invalid Stake, %v", err)
			}
			if v.StakeLedger == nil {
				return fmt.Errorf("unmarshal error: invalid StakeLedger")
			}
			a.StakeLedger, err = LedgerFromBytesList(*v.StakeLedger)
			if err != nil {
				return fmt.Errorf("unmarshal error: %v", err)
			}
		}

		if v.Lending != nil {
			a.Lending, err = LendingConfigFromBytesList(*v.Lending)
			if err != nil {
				return fmt.Errorf("unmarshal error: invalid Lending, %v", err)
			}
			if v.LendingLedger == nil {
				return fmt.Errorf("unmarshal error: invalid LendingLedger")
			}
			a.LendingLedger, err = LedgerFromBytesList(*v.LendingLedger)
			if err != nil {
				return fmt.Errorf("unmarshal error: %v", err)
			}
		}
		a.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindAccount")
}

func (a *Account) Marshal() ([]byte, error) {
	ba := &bindAccount{
		Type:           FromUint8(uint8(a.Type)),
		Nonce:          FromUint64(a.Nonce),
		Balance:        FromUint(a.Balance),
		Threshold:      FromUint8(a.Threshold),
		Keepers:        FromShortIDs(a.Keepers),
		MaxTotalSupply: FromUint(a.MaxTotalSupply),
		NonceTable:     make([][]byte, 0, len(a.NonceTable)),
		Tokens:         make([][]byte, 0, len(a.Tokens)),
	}

	buf := new(bytes.Buffer)
	var err error
	for k, uu := range a.NonceTable {
		if k < a.Timestamp || len(uu) == 0 {
			continue
		}
		if err = WriteUint64s(buf, k, uu...); err != nil {
			return nil, err
		}
		ba.NonceTable = append(ba.NonceTable, buf.Bytes())
		buf.Reset()
	}

	for k, v := range a.Tokens {
		if err = WriteBytesList(buf, k[:], FromUint(v)); err != nil {
			return nil, err
		}
		ba.Tokens = append(ba.Tokens, buf.Bytes())
		buf.Reset()
	}

	if a.Stake != nil {
		list := a.Stake.ToBytesList()
		ba.Stake = &list
		list, err = a.StakeLedger.ToBytesList(buf)
		if err != nil {
			return nil, err
		}
		ba.StakeLedger = &list
	}

	if a.Lending != nil {
		list := a.Lending.ToBytesList()
		ba.Lending = &list
		list, err = a.LendingLedger.ToBytesList(buf)
		if err != nil {
			return nil, err
		}
		ba.LendingLedger = &list
	}

	data, err := accountLDBuilder.Marshal(ba)
	if err != nil {
		return nil, err
	}
	a.raw = data
	return data, nil
}

type bindAccount struct {
	Type           Uint8
	Nonce          Uint64
	Balance        BigUint
	Threshold      Uint8
	Keepers        [][]byte
	NonceTable     [][]byte
	Tokens         [][]byte
	MaxTotalSupply BigUint
	Stake          *[][]byte
	StakeLedger    *[][]byte
	Lending        *[][]byte
	LendingLedger  *[][]byte
}

var accountLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type Account struct {
		Type           Uint8   (rename "t")
		Nonce          Uint64  (rename "n")
		Balance        BigUint (rename "b")
		Threshold      Uint8   (rename "th")
		Keepers        [ID20]  (rename "kp")
		NonceTable     [Bytes] (rename "nt")
		Tokens         [Bytes] (rename "lg")
		MaxTotalSupply BigUint (rename "mts")
		Stake          nullable [Bytes] (rename "st")
		StakeLedger    nullable [Bytes] (rename "stl")
		Lending        nullable [Bytes] (rename "le")
		LendingLedger  nullable [Bytes] (rename "lel")
	}
`
	builder, err := NewLDBuilder("Account", []byte(sch), (*bindAccount)(nil))
	if err != nil {
		panic(err)
	}
	accountLDBuilder = builder
}
