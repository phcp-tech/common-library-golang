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
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptPassword encrypt password, pbkdf2, salt = uuid, with 4096 iterations, keyLen= 32, sha256
func EncryptPassword(password string, salt string) string {
	return base64.StdEncoding.EncodeToString(pbkdf2.Key([]byte(password), []byte(salt), 4096, 32, sha256.New))
}
