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

package dbsqlx_test

import (
	"encoding/json"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
)

// -----------------------------------------------------------------------
// JSONRaw: encoding/json boundary (Marshal/Unmarshal)
// -----------------------------------------------------------------------

func TestJSONRaw_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		j    dbsqlx.JSONRaw
		want string
	}{
		{"object passes through untouched", dbsqlx.JSONRaw(`{"category":"Create feature"}`), `{"category":"Create feature"}`},
		{"array passes through untouched", dbsqlx.JSONRaw(`[1,2,3]`), `[1,2,3]`},
		{"nil marshals to the JSON literal null", nil, "null"},
		{"empty (non-nil) marshals to the JSON literal null", dbsqlx.JSONRaw{}, "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("MarshalJSON = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestJSONRaw_MarshalJSON_StructField guards the actual bug this type
// exists to fix: a plain []byte struct field marshals to a base64 string
// instead of a nested JSON value. JSONRaw must not do that.
func TestJSONRaw_MarshalJSON_StructField(t *testing.T) {
	type row struct {
		Data dbsqlx.JSONRaw `json:"data,omitempty"`
	}
	out, err := json.Marshal(row{Data: dbsqlx.JSONRaw(`{"category":"Create feature"}`)})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	const want = `{"data":{"category":"Create feature"}}`
	if string(out) != want {
		t.Fatalf("Marshal = %s, want %s (looks like it base64-encoded instead)", out, want)
	}
}

func TestJSONRaw_UnmarshalJSON(t *testing.T) {
	var j dbsqlx.JSONRaw
	if err := json.Unmarshal([]byte(`{"category":"Create feature"}`), &j); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(j) != `{"category":"Create feature"}` {
		t.Fatalf("got %s, want the object verbatim (no base64 decoding applied)", j)
	}
}

// TestJSONRaw_UnmarshalJSON_StructField guards the request-side counterpart
// of TestJSONRaw_MarshalJSON_StructField: a client posting a plain JSON
// object for this field must not be rejected the way []byte would reject it
// ("json: cannot unmarshal object into Go struct field ... of type []uint8").
func TestJSONRaw_UnmarshalJSON_StructField(t *testing.T) {
	type row struct {
		Data dbsqlx.JSONRaw `json:"data,omitempty"`
	}
	var r row
	if err := json.Unmarshal([]byte(`{"data":{"category":"Create feature"}}`), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(r.Data) != `{"category":"Create feature"}` {
		t.Fatalf("got %s, want the object verbatim", r.Data)
	}
}

// -----------------------------------------------------------------------
// JSONRaw: database/sql boundary (Scan/Value)
// -----------------------------------------------------------------------

func TestJSONRaw_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    string
		wantErr bool
	}{
		{name: "nil (SQL NULL) becomes a nil JSONRaw", src: nil, want: ""},
		{name: "[]byte is copied in", src: []byte(`{"a":1}`), want: `{"a":1}`},
		{name: "string is copied in", src: `{"a":1}`, want: `{"a":1}`},
		{name: "unsupported type errors", src: 42, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j dbsqlx.JSONRaw
			err := j.Scan(tt.src)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if string(j) != tt.want {
				t.Fatalf("got %q, want %q", j, tt.want)
			}
		})
	}
}

func TestJSONRaw_Value(t *testing.T) {
	tests := []struct {
		name string
		j    dbsqlx.JSONRaw
		want any
	}{
		{"non-empty passes through as []byte", dbsqlx.JSONRaw(`{"a":1}`), []byte(`{"a":1}`)},
		{"nil becomes SQL NULL, not an empty string", nil, nil},
		{"empty (non-nil) becomes SQL NULL, not an empty string", dbsqlx.JSONRaw{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("Value = %#v, want nil (SQL NULL)", got)
				}
				return
			}
			if string(got.([]byte)) != string(tt.want.([]byte)) {
				t.Fatalf("Value = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// JSONRaw: real database round-trip (guards the actual production bug —
// see openTestDB in helpers_test.go for the shared in-memory SQLite fixture)
// -----------------------------------------------------------------------

func TestJSONRaw_DatabaseRoundTrip(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Exec(`CREATE TABLE logs (id INTEGER PRIMARY KEY, data BLOB)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	type logRow struct {
		Id   int64          `db:"id"`
		Data dbsqlx.JSONRaw `db:"data"`
	}

	t.Run("a genuine JSON payload round-trips intact", func(t *testing.T) {
		if _, err := db.Exec(`INSERT INTO logs (id, data) VALUES (?, ?)`, 1, dbsqlx.JSONRaw(`{"category":"Create feature"}`)); err != nil {
			t.Fatalf("insert: %v", err)
		}
		var got logRow
		if err := db.Get(&got, `SELECT id, data FROM logs WHERE id = ?`, 1); err != nil {
			t.Fatalf("select: %v", err)
		}
		if string(got.Data) != `{"category":"Create feature"}` {
			t.Fatalf("got %s", got.Data)
		}
	})

	t.Run("a NULL column scans to a nil JSONRaw without error", func(t *testing.T) {
		if _, err := db.Exec(`INSERT INTO logs (id, data) VALUES (?, NULL)`, 2); err != nil {
			t.Fatalf("insert: %v", err)
		}
		var got logRow
		if err := db.Get(&got, `SELECT id, data FROM logs WHERE id = ?`, 2); err != nil {
			t.Fatalf("select: %v", err)
		}
		if got.Data != nil {
			t.Fatalf("expected nil, got %s", got.Data)
		}
	})

	t.Run("an empty (non-nil) JSONRaw binds as NULL, not an empty string", func(t *testing.T) {
		// Reproduces a client POSTing `"data": ""` for a JSONRaw field: it
		// decodes to a non-nil, zero-length value — exactly this case.
		if _, err := db.Exec(`INSERT INTO logs (id, data) VALUES (?, ?)`, 3, dbsqlx.JSONRaw{}); err != nil {
			t.Fatalf("insert should not fail the way an empty-string bind would against a real json/jsonb column: %v", err)
		}
		var got logRow
		if err := db.Get(&got, `SELECT id, data FROM logs WHERE id = ?`, 3); err != nil {
			t.Fatalf("select: %v", err)
		}
		if got.Data != nil {
			t.Fatalf("expected the empty value to have been stored as NULL, got %s", got.Data)
		}
	})
}
