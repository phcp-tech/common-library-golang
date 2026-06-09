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

package validator

import "testing"

// ---------------------------------------------------------------------------
// validator.go – Validator (floatin, choicelist custom validators)
// ---------------------------------------------------------------------------

func TestValidator_Floatin_Valid(t *testing.T) {
	type S struct {
		F float64 `validate:"floatin=1.0 2.5 3.7"`
	}
	v := Validator()
	if err := v.Struct(S{F: 2.5}); err != nil {
		t.Errorf("Validator floatin valid: unexpected error: %v", err)
	}
}

func TestValidator_Floatin_Invalid(t *testing.T) {
	type S struct {
		F float64 `validate:"floatin=1.0 2.5 3.7"`
	}
	v := Validator()
	if err := v.Struct(S{F: 9.9}); err == nil {
		t.Error("Validator floatin invalid: expected error, got nil")
	}
}

func TestValidator_ChoiceList_Valid(t *testing.T) {
	type S struct {
		C string `validate:"choicelist=A B C"`
	}
	v := Validator()
	if err := v.Struct(S{C: "A,B"}); err != nil {
		t.Errorf("Validator choicelist valid: unexpected error: %v", err)
	}
}

func TestValidator_ChoiceList_Invalid(t *testing.T) {
	type S struct {
		C string `validate:"choicelist=A B C"`
	}
	v := Validator()
	if err := v.Struct(S{C: "A,X"}); err == nil {
		t.Error("Validator choicelist invalid: expected error, got nil")
	}
}

func TestValidator_StandardTags(t *testing.T) {
	type S struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=130"`
	}
	v := Validator()
	if err := v.Struct(S{Email: "test@example.com", Age: 25}); err != nil {
		t.Errorf("Validator standard tags: unexpected error: %v", err)
	}
	if err := v.Struct(S{Email: "bad-email", Age: 25}); err == nil {
		t.Error("Validator standard tags: expected error for bad email, got nil")
	}
}

func TestValidator_Singleton(t *testing.T) {
	v1 := Validator()
	v2 := Validator()
	if v1 != v2 {
		t.Error("Validator should return the same singleton instance on every call")
	}
}
