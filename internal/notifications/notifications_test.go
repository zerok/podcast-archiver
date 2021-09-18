package notifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSample(t *testing.T) {
	ctx := context.Background()
	n, err := NewFromEnv()
	require.NoError(t, err)
	if n == nil {
		t.SkipNow()
		return
	}
	require.NoError(t, err)
	require.NoError(t, n.Login(ctx))
	require.NoError(t, n.JoinRoom(ctx))
	require.NoError(t, n.Send(ctx, "hello"))
}
