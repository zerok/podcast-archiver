// Copyright 2017, Horst Gutmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
