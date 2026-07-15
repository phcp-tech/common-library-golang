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

package errors

import "github.com/phcp-tech/common-library-golang/dto"

// API response code constants used as the top-level status indicator in JSON responses.
const (
	API_CODE_SUCCESS = 0 // API_CODE_SUCCESS indicates the request was processed successfully.
	API_CODE_FAILED  = 1 // API_CODE_FAILED indicates the request failed.
)

// Pre-defined ResponseMessage error values for common token and server errors.
var (
	// ErrorNotFound is returned when the requested resource does not exist.
	ErrorNotFound = &dto.ResponseMessage{Code: 404, Message: "record not found", Data: nil}
	// ErrorNoToken is returned when no token is present in the request.
	ErrorNoToken = &dto.ResponseMessage{Code: 400, Message: "token does not define", Data: nil}
	// ErrorBadToken is returned when the provided token is malformed or invalid.
	ErrorBadToken = &dto.ResponseMessage{Code: 400, Message: "token error", Data: nil}
	// ErrorExpiredToken is returned when the provided token has expired.
	ErrorExpiredToken = &dto.ResponseMessage{Code: 401, Message: "token expired", Data: nil}
	// ErrorPermissionToken is returned when the token lacks the required permission.
	ErrorPermissionToken = &dto.ResponseMessage{Code: 403, Message: "token has no permission", Data: nil}
	// ErrorInternalServer is returned for unexpected server-side errors.
	ErrorInternalServer = &dto.ResponseMessage{Code: 500, Message: "internal server error", Data: nil}
	// ErrorDatabase is returned when a database operation fails.
	ErrorDatabase = &dto.ResponseMessage{Code: 500, Message: "database error", Data: nil}
	// ErrorVersionConflict is returned when an update carries a stale version:
	// another user has saved a change to the same field since this caller last
	// read it. Defined here (rather than as a package-level sentinel in
	// service, à la ErrProductHasFeatures) because infra/dao must be able to
	// return it too, and infra/dao cannot import service. Callers compare via
	// err.Error() == ErrorVersionConflict.Message, matching the existing
	// errorcode.ErrorNotFound.Message convention used across adapter/infra/dao.
	ErrorVersionConflict = &dto.ResponseMessage{Code: 409, Message: "content has been modified by another user", Data: nil}
)
