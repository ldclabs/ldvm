// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxCreateStakeAccount(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateStakeAccount{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)
}
