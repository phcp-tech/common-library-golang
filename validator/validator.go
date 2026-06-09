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

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// Validator is a lazily-initialised, package-level singleton of *validator.Validate.
// It caches struct metadata for efficiency and registers the custom validation tags
// "floatin" (checks a float field against a whitelist of allowed values) and
// "choicelist" (checks that every comma-separated token is in an allowed set).
var Validator = sync.OnceValue(func() *validator.Validate {
	validator := validator.New()
	validator.RegisterValidation("floatin", floatInValidator)
	validator.RegisterValidation("choicelist", choiceListValidator)
	return validator
})

// custom validator for floatin
func floatInValidator(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.Float32 && field.Kind() != reflect.Float64 {
		return false
	}

	param := fl.Param()
	allowedValues := strings.Fields(param)

	fieldValue := field.Float()
	for _, val := range allowedValues {
		allowedFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}

		if fieldValue == allowedFloat {
			return true
		}
	}
	return false
}

// custom validator for choiceListValidator
func choiceListValidator(fl validator.FieldLevel) bool {
	// Get the string input by the user
	value := fl.Field().String()
	inputs := strings.Split(value, ",")

	// Use ' ' as a separator, because the default label syntax of Validator uses ',' to separate multiple rules
	// Get validator tag parameters, such as "A B".
	param := fl.Param()
	allowedSet := make(map[string]struct{}, 10)
	for _, val := range strings.Split(param, " ") {
		allowedSet[strings.TrimSpace(val)] = struct{}{}
	}

	// Checks if each value is in allowedSet
	for _, item := range inputs {
		item = strings.TrimSpace(item)
		if _, ok := allowedSet[item]; !ok {
			return false
		}
	}
	return true
}
