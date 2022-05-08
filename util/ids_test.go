// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

// DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i
const address1 = "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"

// 7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6
const address2 = "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"

func TestEthID(t *testing.T) {
	assert := assert.New(t)

	id1, err := EthIDFromString(address1)
	assert.Nil(err)
	assert.Equal(Signer1.Address(), id1)

	id2, err := EthIDFromString("8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.Nil(err)
	assert.Equal(id1, id2)

	id3, err := EthIDFromString("DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i")
	assert.Nil(err)
	assert.Equal(id1, id3)

	id, err := EthIDFromString(address2)
	assert.Nil(err)
	assert.Equal(Signer2.Address(), id)

	cbordata, err := cbor.Marshal(ids.ID{1, 2, 3})
	assert.Nil(err)
	var id4 EthID
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id4), "invalid length bytes: 32")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id4))
	assert.Equal(id, id4)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"`, string(data))

	eids := make([]EthID, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6",
		"",
		null
	]`), &eids)
	assert.Nil(err)

	assert.Equal(5, len(eids))
	assert.Equal(id, eids[0])
	assert.Equal(id, eids[1])
	assert.Equal(id, eids[2])
	assert.Equal(EthIDEmpty, eids[3])
	assert.Equal(EthIDEmpty, eids[4])

	ptrIDs := make([]*EthID, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6",
		"",
		null
	]`), &ptrIDs)
	assert.Nil(err)

	assert.Equal(5, len(eids))
	assert.Equal(id, *ptrIDs[0])
	assert.Equal(id, *ptrIDs[1])
	assert.Equal(id, *ptrIDs[2])
	assert.Equal(EthIDEmpty, *ptrIDs[3])
	assert.Nil(ptrIDs[4])

	id, err = EthIDFromString("")
	assert.Nil(err)
	assert.Equal(EthIDEmpty, id)
}

func TestModelID(t *testing.T) {
	assert := assert.New(t)

	mid := "LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := ModelIDFromString(mid)
	assert.Nil(err)

	cbordata, err := cbor.Marshal(ids.ID{1, 2, 3})
	assert.Nil(err)
	var id2 ModelID
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid length bytes: 32")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"`, string(data))

	mids := make([]ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	assert.Nil(err)

	assert.Equal(3, len(mids))
	assert.Equal(id, mids[0])
	assert.Equal(ModelIDEmpty, mids[1])
	assert.Equal(ModelIDEmpty, mids[2])

	ptrMIDs := make([]*ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(ModelIDEmpty, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = ModelIDFromString("")
	assert.Nil(err)
	assert.Equal(ModelIDEmpty, id)
}

func TestDataID(t *testing.T) {
	assert := assert.New(t)

	mid := "LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := DataIDFromString(mid)
	assert.Nil(err)

	cbordata, err := cbor.Marshal(ids.ID{1, 2, 3})
	assert.Nil(err)
	var id2 DataID
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid length bytes: 32")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"`, string(data))

	mids := make([]DataID, 0)
	err = json.Unmarshal([]byte(`[
		"LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	assert.Nil(err)

	assert.Equal(3, len(mids))
	assert.Equal(id, mids[0])
	assert.Equal(DataIDEmpty, mids[1])
	assert.Equal(DataIDEmpty, mids[2])

	ptrMIDs := make([]*DataID, 0)
	err = json.Unmarshal([]byte(`[
		"LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(DataIDEmpty, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = DataIDFromString("")
	assert.Nil(err)
	assert.Equal(DataIDEmpty, id)
}

func TestTokenSymbol(t *testing.T) {
	assert := assert.New(t)

	token := "LDC"
	id, err := NewToken(token)
	assert.Nil(err)

	assert.Equal(
		TokenSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal(ids.ID{'L', 'D', 'C'})
	assert.Nil(err)
	var id2 TokenSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid length bytes: 32")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"LDC"`, string(data))

	id, err = NewToken("")
	assert.Nil(err)
	assert.Equal(NativeToken, id)

	type testCase struct {
		shouldErr bool
		symbol    string
		token     TokenSymbol
	}
	tcs := []testCase{
		{shouldErr: false, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: false, symbol: "LD",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 'L', 'D',
			}},
		{shouldErr: false, symbol: "USD",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "U1D",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'U', '1', 'D',
			}},
		{shouldErr: false, symbol: "USD1",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "L012345678",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'L', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 'D', 0,
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 'C',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 0, 'c',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
			}},
		{shouldErr: true, symbol: "LDc",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "L_C",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "L C",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1LDC",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1234567890",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "LD‍C", // with Zero Width Joiner
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
	}
	for _, c := range tcs {
		switch {
		case c.shouldErr:
			assert.Equal("", c.token.String())
			if c.token != NativeToken {
				assert.False(c.token.Valid())
			}
			if c.symbol != "" {
				_, err := NewToken(c.symbol)
				assert.NotNil(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := NewToken(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}

func FuzzTokenSymbol(f *testing.F) {
	for _, seed := range []string{
		"",
		"AVAX",
		"abc",
		"A100",
		"ABCDEFGHIJKLMNOPQRST",
	} {
		f.Add(seed)
	}
	counter := 0
	f.Fuzz(func(t *testing.T, in string) {
		id, err := NewToken(in)
		switch {
		case err == nil:
			counter++
			assert.Equal(t, in, id.String())
		default:
		}
	})
	assert.True(f, counter > 0)
}

func TestStakeSymbol(t *testing.T) {
	assert := assert.New(t)

	token := "@LDC"
	id, err := NewStake(token)
	assert.Nil(err)

	assert.Equal(
		StakeSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, '@', 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal(ids.ID{'@', 'L', 'D', 'C'})
	assert.Nil(err)
	var id2 StakeSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid length bytes: 32")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"@LDC"`, string(data))

	id, err = NewStake("")
	assert.Nil(err)
	assert.Equal(StakeEmpty, id)

	type testCase struct {
		shouldErr bool
		symbol    string
		token     StakeSymbol
	}
	tcs := []testCase{
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: false, symbol: "@D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, '@', 'D',
			}},
		{shouldErr: false, symbol: "@USD",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '@', 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "@1D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '@', '1', 'D',
			}},
		{shouldErr: false, symbol: "@USD1",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, '@', 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "@012345678",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'@', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 'D', 0,
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, '@',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '@', 'L', 'D', 'c',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				'@', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'l',
			}},
		{shouldErr: true, symbol: "@LDc",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "@L_C",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "@L C",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1LDC",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1234567890",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "@LD‍C", // with Zero Width Joiner
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
	}
	for _, c := range tcs {
		switch {
		case c.shouldErr:
			assert.Equal("", c.token.String())
			assert.False(c.token.Valid())
			if c.symbol != "" {
				_, err := NewStake(c.symbol)
				assert.NotNil(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := NewStake(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}

func TestEthIDToStakeSymbol(t *testing.T) {
	assert := assert.New(t)

	ldc, err := NewStake("@LDC")
	assert.Nil(err)
	ids := []EthID{
		EthID(ldc),
		EthIDEmpty,
		Signer1.Address(),
		Signer2.Address(),
	}
	ss := EthIDToStakeSymbol(ids...)
	assert.Equal(ldc, ss[0])
	assert.Equal("@LDC", ss[0].String())
	assert.Equal("@6NUDZHR5VGT7SA4XOZZ", ss[1].String())
	assert.Equal("@6NUDZHR5VGT7SA4XOZZ", string(ss[1][:]))
	assert.Equal("@BLDQHR4QOJZMNIC5Q5U", ss[2].String())
	assert.Equal("@BLDQHR4QOJZMNIC5Q5U", string(ss[2][:]))
	assert.Equal("@GWLGDBWNPCOAN55PCUX", ss[3].String())
	assert.Equal("@GWLGDBWNPCOAN55PCUX", string(ss[3][:]))
}
