// Copyright 2020-2021 Dolthub, Inc.
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

package expression

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vela-security/vela-chameleon/mysql/sql"
)

func eval(t *testing.T, e sql.Expression, row sql.Row) interface{} {
	t.Helper()
	v, err := e.Eval(sql.NewEmptyContext(), row)
	require.NoError(t, err)
	return v
}

func TestIsUnary(t *testing.T) {
	require := require.New(t)
	require.True(IsUnary(NewNot(nil)))
	require.False(IsUnary(NewAnd(nil, nil)))
}

func TestIsBinary(t *testing.T) {
	require := require.New(t)
	require.False(IsBinary(NewNot(nil)))
	require.True(IsBinary(NewAnd(nil, nil)))
}
