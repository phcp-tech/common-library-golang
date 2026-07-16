// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbsqlx

import (
	"database/sql/driver"
	"fmt"
)

// JSONRaw holds an arbitrary JSON value (object, array, string, ...) for a
// database column of type json/jsonb that's also exposed at a JSON API
// boundary (bound from a client request body, or marshaled into a response).
//
// A plain []byte field is the wrong type for that: it round-trips correctly
// through a database column (sqlx binds/scans []byte directly), but
// encoding/json treats []byte specially — a []byte struct field marshals to
// a base64 string, and only unmarshals from one, not from a JSON object. A
// client sending a normal JSON object for that field fails with "json:
// cannot unmarshal object into Go struct field ... of type []uint8"; a
// server returning one instead returns a meaningless base64 string that
// silently isn't the JSON object callers expect.
//
// JSONRaw fixes both directions: MarshalJSON/UnmarshalJSON pass the bytes
// through untouched, so a JSON object stays a JSON object on the wire, both
// into and out of the API — while Scan/Value make it bind/scan against a
// json/jsonb database column exactly like []byte already did.
//
// It also closes a second, sharper edge: a caller that never touches the
// field ends up with a nil JSONRaw, but a client that explicitly POSTs
// `"data": ""` for a []byte field decodes (via base64) to a non-nil, empty
// slice — indistinguishable from "no data" by intent, but not by a `== nil`
// guard. Passed straight through to an INSERT/UPDATE, an empty (but
// non-nil) value is bound as an empty string, which every Postgres
// json/jsonb column rejects as invalid input: "invalid input syntax for
// type json" (SQLSTATE 22P02). JSONRaw's Value() treats any zero-length
// value as SQL NULL instead, so this can't happen regardless of whether the
// call site remembered to guard against it.
//
// Deliberately not encoding/json.RawMessage: some of this library's
// consumers build with GOEXPERIMENT=jsonv2, under which RawMessage no
// longer satisfies database/sql's generic []byte-scan fallback — Scan then
// errors ("unsupported Scan ... into type *jsontext.Value") for both NULL
// and non-NULL driver values. JSONRaw sidesteps that fallback entirely by
// implementing sql.Scanner/driver.Valuer itself.
//
// Also deliberately not gorm.io/datatypes.JSON, which already solves the
// same three problems: reusing it would pull the entire GORM module into
// every consumer of this package, including the sqlx-only, no-ORM services
// that migrated away from GORM specifically to avoid that dependency.
type JSONRaw []byte

// MarshalJSON implements json.Marshaler, emitting the held bytes as-is —
// callers marshaling a struct with a JSONRaw field get the field back as a
// real JSON value, not a base64 string. An empty/nil JSONRaw marshals to
// the JSON literal null (mirrors encoding/json.RawMessage's own behavior).
func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON implements json.Unmarshaler, storing data as-is — a JSON
// object/array in the request body lands in the field verbatim, with no
// base64 decoding step required (or accepted) from the caller.
func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("dbsqlx.JSONRaw: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Scan implements sql.Scanner. A SQL NULL becomes a nil JSONRaw rather than
// an error — the same "no data set" state a genuinely absent value has.
func (j *JSONRaw) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
		return nil
	case string:
		*j = JSONRaw(v)
		return nil
	default:
		return fmt.Errorf("dbsqlx.JSONRaw: unsupported Scan type %T", src)
	}
}

// Value implements driver.Valuer. A zero-length JSONRaw becomes SQL NULL
// rather than an empty string — see the type doc comment for why this
// matters: an empty string is not valid input for a json/jsonb column.
func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}
