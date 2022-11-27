// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Session struct {
	UID string
}

func (s *Session) Valid() bool {
	return s != nil
}

type Session2 Session

func (s *Session2) Valid() bool {
	return s != nil
}

func TestCtx(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		assert := assert.New(t)

		v0 := Session{"a"}
		ctx := CtxWith(context.Background(), &v0)

		v1 := CtxValue[Session](context.Background())
		assert.True(v1 == nil)

		v1 = CtxValue[Session](ctx)
		require.NotNil(t, v1)
		assert.Equal(v0, *v1)

		*v1 = Session{"abc"}
		assert.Equal(v0, *v1)

		v2 := CtxValue[Session](ctx)
		assert.Equal(v0, *v2)

		called := false
		DoIfCtxValueValid(ctx, func(v *Session2) { called = true })
		assert.False(called)

		DoIfCtxValueValid(ctx, func(v *Session) { called = true })
		assert.True(called)

		v3 := CtxValue[Session2](ctx)
		assert.True(v3 == nil)
	})
}
