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

package email

// Message is an outbound email: sender, recipients and content. Attachments
// are not supported. The validate tags are for validator.Validator() when the
// message comes from a request body.
type Message struct {
	From        []string `json:"from,omitempty" validate:"omitempty,min=1,max=2"`          // [0] sender address, [1] optional display name; empty = the sender of Config
	To          []string `json:"to,omitempty" validate:"required,min=1,dive,email"`        // primary recipient addresses
	Cc          []string `json:"cc,omitempty" validate:"omitempty,dive,email"`             // carbon-copy recipient addresses
	Bcc         []string `json:"bcc,omitempty" validate:"omitempty,dive,email"`            // blind carbon-copy recipient addresses
	Subject     string   `json:"subject,omitempty" validate:"omitempty,min=0,max=500"`     // subject line
	Body        string   `json:"body,omitempty"`                                           // body text
	ContentType string   `json:"contentType,omitempty" validate:"omitempty,min=0,max=100"` // text/plain or text/html (default); parameters such as charset are ignored
}
