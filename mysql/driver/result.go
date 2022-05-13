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

package driver

import (
	"errors"
	"io"

	"github.com/vela-security/vela-chameleon/mysql/sql"
)

func getOKResult(ctx *sql.Context, rows sql.RowIter) (sql.OkResult, bool, error) {
	var okr sql.OkResult
	var found bool
	for !found {
		row, err := rows.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return okr, found, err
		}

		if len(row) != 1 {
			continue
		}

		okr, found = row[0].(sql.OkResult)
	}

	err := rows.Close(ctx)
	return okr, found, err
}

// Result is the result of a query execution.
type Result struct {
	result sql.OkResult
}

// LastInsertId returns the database's auto-generated ID
// after, for example, an INSERT into a table with primary
// key.
//
// NOT IMPLEMENTED
func (r *Result) LastInsertId() (int64, error) {
	return 0, errors.New("sql.Result.LastInsertId is not implemented")
}

// RowsAffected returns the number of rows affected by the
// query.
func (r *Result) RowsAffected() (int64, error) {
	return int64(r.result.RowsAffected), nil
}

// ResultNotFound is returned when a row iterator does not return a result.
type ResultNotFound struct{}

// LastInsertId returns an error
func (r *ResultNotFound) LastInsertId() (int64, error) {
	return 0, errors.New("no result")
}

// RowsAffected returns an error
func (r *ResultNotFound) RowsAffected() (int64, error) {
	return 0, errors.New("no result")
}
