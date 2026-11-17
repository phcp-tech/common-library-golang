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

package encrypt

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestEncryptPassword_KnownVector(t *testing.T) {
	// PBKDF2-HMAC-SHA256, P="password", S="salt", c=4096, dkLen=32.
	want, _ := hex.DecodeString("c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a")
	if got := EncryptPassword("password", "salt"); got != base64.StdEncoding.EncodeToString(want) {
		t.Errorf("EncryptPassword = %q, want %q", got, base64.StdEncoding.EncodeToString(want))
	}
}

func TestEncryptPassword_SaltAndPasswordChangeResult(t *testing.T) {
	base := EncryptPassword("secret", "salt-1")
	if base != EncryptPassword("secret", "salt-1") {
		t.Error("same input: want the same hash")
	}
	if base == EncryptPassword("secret", "salt-2") {
		t.Error("different salt: want a different hash")
	}
	if base == EncryptPassword("secret2", "salt-1") {
		t.Error("different password: want a different hash")
	}
}

func TestEncryptPassword_OutputFormat(t *testing.T) {
	for _, c := range []struct{ password, salt string }{
		{"secret", "salt"},
		{"", ""}, // empty input is allowed and still gives a full-length hash
		{"密码 with spaces", "0190a3b2-7c4e-7000-8000-000000000000"},
	} {
		got := EncryptPassword(c.password, c.salt)
		raw, err := base64.StdEncoding.DecodeString(got)
		if err != nil || len(raw) != 32 {
			t.Errorf("(%q, %q): want 32 bytes of standard base64, got %q (%d bytes, err %v)", c.password, c.salt, got, len(raw), err)
		}
		if len(got) != 44 {
			t.Errorf("(%q, %q): encoded length = %d, want 44", c.password, c.salt, len(got))
		}
	}
}
