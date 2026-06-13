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

package dto

// LoginUser holds the claims embedded inside a JWT token for an authenticated
// user. Fields such as UserId, Username, ProductId, and Roles must not use
// omitempty so that zero values are always serialised.
type LoginUser struct {
	Token     string   `json:"token,omitempty"`     // stored token used for inter-service REST calls; not included when generating a new token
	TokenType string   `json:"tokenType,omitempty"` // token purpose: "access" or "refresh"
	UserId    uint64   `json:"userId"`
	Username  string   `json:"username"`
	ProductId uint64   `json:"productId"`
	Roles     []string `json:"roles"`
}

// ResponseMessage is the standard response envelope returned by all API
// endpoints. Code 0 indicates success; any other value indicates failure.
type ResponseMessage struct {
	Code    int    `json:"code"`              // 0: success; others: failed
	Message string `json:"message,omitempty"` // omit in success response
	Data    any    `json:"data,omitempty"`    // omit in failed response
}

// DataListResp is the standard paginated list response envelope.
// Both Total and List must not be omitted even when zero/nil so that
// the frontend can always rely on their presence.
type DataListResp struct {
	Total int         `json:"total"` // total number of matching records; must not be omitted
	List  interface{} `json:"list"`  // slice of result items; must not be omitted
}

type ResourceResp struct {
	Id     int    `json:"id"`
	Name   string `json:"name,omitempty"`
	Status int    `json:"status,omitempty"`
}

// PageParameter holds pagination and sorting parameters for list queries.
// Page and Limit must always be present (non-omitempty) so that the
// frontend can rely on their presence in the JSON payload.
type PageParameter struct {
	Page      int    `json:"page"`  // current page number (1-based); must not be omitted
	Limit     int    `json:"limit"` // number of records per page; must not be omitted
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty" validate:"omitempty,oneofci=ASC DESC"` // sort direction: ASC or DESC
	Charset   string `json:"charset,omitempty"`
}
