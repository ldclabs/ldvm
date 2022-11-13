// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

func TestDataID(t *testing.T) {
	assert := assert.New(t)

	did := "CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ"
	addr, err := AddressFromStr(address1)
	assert.Nil(err)
	id := DataID(sha3.Sum256(addr[:]))
	assert.Equal(did, id.String())

	cbordata := encoding.MustMarshalCBOR([20]byte{1, 2, 3})
	var id2 DataID
	assert.ErrorContains(encoding.UnmarshalCBOR(cbordata, &id2), "invalid bytes length")

	cbordata = encoding.MustMarshalCBOR(id)
	assert.Nil(encoding.UnmarshalCBOR(cbordata, &id2))
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
	assert.Equal(EmptyDataID, mids[1])
	assert.Equal(EmptyDataID, mids[2])

	ptrMIDs := make([]*DataID, 0)
	err = json.Unmarshal([]byte(`[
		"CscDx5BycsagXYdpwTk8v7eQk4NKPzreiYRfP_qLqwzDe_zZ",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(EmptyDataID, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = DataIDFromStr("")
	assert.Nil(err)
	assert.Equal(EmptyDataID, id)
}
