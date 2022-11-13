// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
)

func TestModelID(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye", EmptyModelID.String())
	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah", ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}.String())
	assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw", ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 2,
	}.String())

	mid := "jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"
	addr, err := AddressFromStr(address1)
	assert.Nil(err)
	assert.Equal(mid, ModelID(addr).String())

	id, err := ModelIDFromStr(mid)
	assert.Nil(err)
	assert.Equal(ModelID(addr), id)

	cbordata := encoding.MustMarshalCBOR([32]byte{1, 2, 3})
	var id2 ModelID
	assert.ErrorContains(encoding.UnmarshalCBOR(cbordata, &id2), "invalid bytes length")

	cbordata = encoding.MustMarshalCBOR(id)
	assert.Nil(encoding.UnmarshalCBOR(cbordata, &id2))
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
	assert.Equal(EmptyModelID, mids[1])
	assert.Equal(EmptyModelID, mids[2])

	ptrMIDs := make([]*ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH",
		"",
		null
	]`), &ptrMIDs)
	assert.Nil(err)

	assert.Equal(3, len(ptrMIDs))
	assert.Equal(id, *ptrMIDs[0])
	assert.Equal(EmptyModelID, *ptrMIDs[1])
	assert.Nil(ptrMIDs[2])

	id, err = ModelIDFromStr("")
	assert.Nil(err)
	assert.Equal(EmptyModelID, id)
}
