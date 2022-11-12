// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

const address1 = "0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc"

const address2 = "0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"

func TestAddressID(t *testing.T) {
	assert := assert.New(t)

	addr1, err := AddressFrom(address1)
	assert.Nil(err)
	assert.Equal(address1, addr1.String())

	addr1b, err := AddressFrom("8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.Nil(err)
	assert.Equal(addr1, addr1b)

	addr2, err := AddressFrom(address2)
	assert.Nil(err)
	assert.Equal(address2, addr2.String())

	cbordata, err := cbor.Marshal(ids.ID{1, 2, 3})
	assert.Nil(err)
	var addr2b Address
	assert.ErrorContains(cbor.Unmarshal(cbordata, &addr2b), "invalid bytes length")

	cbordata, err = cbor.Marshal(addr2)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &addr2b))
	assert.Equal(addr2, addr2b)

	data, err := json.Marshal(addr2)
	assert.Nil(err)
	assert.Equal(`"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"`, string(data))

	eids := make(IDList[Address], 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
		"",
		null
	]`), &eids)
	assert.Nil(err)

	assert.Equal(4, len(eids))
	assert.Equal(addr2, eids[0])
	assert.Equal(addr2, eids[1])
	assert.Equal(AddressEmpty, eids[2])
	assert.Equal(AddressEmpty, eids[3])

	ptrIDs := make([]*Address, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
		"",
		null
	]`), &ptrIDs)
	assert.Nil(err)

	assert.Equal(4, len(eids))
	assert.Equal(addr2, *ptrIDs[0])
	assert.Equal(addr2, *ptrIDs[1])
	assert.Equal(AddressEmpty, *ptrIDs[2])
	assert.Nil(ptrIDs[3])

	addr2, err = AddressFrom("")
	assert.Nil(err)
	assert.Equal(AddressEmpty, addr2)
}

func TestModelID(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye", ModelIDEmpty.String())
	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah", ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}.String())
	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw", ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 2,
	}.String())

	mid := "jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"
	addr, err := AddressFrom(address1)
	assert.Nil(err)
	assert.Equal(mid, ModelID(addr).String())

	id, err := ModelIDFrom(mid)
	assert.Nil(err)
	assert.Equal(ModelID(addr), id)

	cbordata := MustMarshalCBOR(ids.ID{1, 2, 3})
	var id2 ModelID
	assert.ErrorContains(UnmarshalCBOR(cbordata, &id2), "invalid bytes length")

	cbordata = MustMarshalCBOR(id)
	assert.Nil(UnmarshalCBOR(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"`, string(data))

	mids := make([]ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH",
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
		"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(ModelIDEmpty, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = ModelIDFrom("")
	assert.Nil(err)
	assert.Equal(ModelIDEmpty, id)
}

func TestDataID(t *testing.T) {
	assert := assert.New(t)

	did := "CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ"
	addr, err := AddressFrom(address1)
	assert.Nil(err)
	id := DataID(sha3.Sum256(addr[:]))
	assert.Equal(did, id.String())

	cbordata := MustMarshalCBOR(ids.ShortID{1, 2, 3})
	var id2 DataID
	assert.ErrorContains(UnmarshalCBOR(cbordata, &id2), "invalid bytes length")

	cbordata = MustMarshalCBOR(id)
	assert.Nil(UnmarshalCBOR(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ"`, string(data))

	mids := make(IDList[DataID], 0)
	err = json.Unmarshal([]byte(`[
		"CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ",
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
		"CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(DataIDEmpty, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = DataIDFrom("")
	assert.Nil(err)
	assert.Equal(DataIDEmpty, id)
}

func TestTokenSymbol(t *testing.T) {
	assert := assert.New(t)

	token := "$LDC"
	id, err := TokenFrom(token)
	assert.Nil(err)

	assert.Equal(
		TokenSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, '$', 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal(ids.ID{'$', 'L', 'D', 'C'})
	assert.Nil(err)
	var id2 TokenSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid bytes length")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"$LDC"`, string(data))

	id, err = TokenFrom("")
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
		{shouldErr: false, symbol: "$D",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, '$', 'D',
			}},
		{shouldErr: false, symbol: "$USD",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '$', 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "$1D",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '$', '1', 'D',
			}},
		{shouldErr: false, symbol: "$USD1",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, '$', 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "$012345678",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'$', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: false, symbol: "$ABCDEFGHIJ012345678",
			token: TokenSymbol{
				'$', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
				'J', '0', '1', '2', '3', '4', '5', '6', '7', '8',
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
				0, 0, 0, 0, 0, 0, 0, 0, 0, '$',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '$', 0, 'c',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
			}},
		{shouldErr: true, symbol: "$LDc",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "$L_C",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "$L C",
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
		{shouldErr: true, symbol: "$LD\u200dC", // with Zero Width Joiner
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
				_, err := TokenFrom(c.symbol)
				assert.Error(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := TokenFrom(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}

func FuzzTokenSymbol(f *testing.F) {
	for _, seed := range []string{
		"",
		"$AVAX",
		"abc",
		"$A100",
		"$ABCDEFGHIJKLMNOPQRST",
	} {
		f.Add(seed)
	}
	counter := 0
	f.Fuzz(func(t *testing.T, in string) {
		id, err := TokenFrom(in)
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

	token := "#LDC"
	id, err := StakeFrom(token)
	assert.Nil(err)

	assert.Equal(
		StakeSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, '#', 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal(ids.ID{'#', 'L', 'D', 'C'})
	assert.Nil(err)
	var id2 StakeSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid bytes length")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"#LDC"`, string(data))

	id, err = StakeFrom("")
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
		{shouldErr: false, symbol: "#D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, '#', 'D',
			}},
		{shouldErr: false, symbol: "#USD",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '#', 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "#1D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '#', '1', 'D',
			}},
		{shouldErr: false, symbol: "#USD1",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, '#', 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "#012345678",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'#', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 'D', 0,
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, '#',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '#', 'L', 'D', 'c',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				'#', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'l',
			}},
		{shouldErr: true, symbol: "#LDc",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "#L_C",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "#L C",
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
		{shouldErr: true, symbol: "#LD\u200dC", // with Zero Width Joiner
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
	}
	for _, c := range tcs {
		switch {
		case c.shouldErr:
			assert.Equal("", c.token.String())
			if c.symbol != "" {
				_, err := StakeFrom(c.symbol)
				assert.Error(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := StakeFrom(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}
