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

// Package component provides JWT token lifecycle integration for bootstrap.
package component

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/token"
)

// loadFromEnv reads JWT configuration from the koanf env singleton and
// initialises the package-level token signing secrets.
// Configuration keys:
//
//	jwt.issuer             — JWT issuer identifier (required — see below)
//	jwt.access.secretcode  — access token signing key (required)
//	jwt.refresh.secretcode — refresh token signing key (required only if the
//	                         service's own config declares it — see below)
//
// jwt.issuer has no per-service default (deliberately — an earlier version of
// this component fell back to app.name, which is exactly the wrong default:
// ParseToken/ParseRefreshToken reject a token whose Issuer claim doesn't match
// the validating service's own issuer, so this value must be identical across
// every service that issues or validates the same tokens — app.name is
// unique per service, so that fallback silently guaranteed every service
// would mint/expect a different issuer and break cross-service auth the
// moment two services actually exchanged a token). Init fails if jwt.issuer
// is empty or left at its placeholder value, exactly like jwt.access.secretcode.
//
// Init fails if jwt.access.secretcode is empty or left at its placeholder value —
// every service registering token.Authenticate() depends on this secret being real.
//
// jwt.refresh.secretcode is treated as required only when it's explicitly left
// at the placeholder value: a service whose config never mentions it at all
// (env.Env() then returns "") never calls CreateRefreshToken/ParseRefreshToken,
// so there's nothing to validate — but a service like common-user, whose
// config/app.toml does declare refresh.secretcode = "TOBE_REPLACED", is
// explicitly saying it needs one, so Init fails here rather than letting that
// service start and only discover the unconfigured secret on a user's first
// login/refresh call.
func loadFromEnv() error {
	issuer := env.Env().String("jwt.issuer")
	access := env.Env().String("jwt.access.secretcode")
	refresh := env.Env().String("jwt.refresh.secretcode")

	if !token.IsUsableSecret(issuer) {
		return fmt.Errorf("token: jwt.issuer is not configured (got %q) — this value must be identical across every service that issues or validates the same tokens", issuer)
	}
	if !token.IsUsableSecret(access) {
		return fmt.Errorf("token: jwt.access.secretcode is not configured (got %q)", access)
	}
	if refresh == token.PlaceholderSecret {
		return fmt.Errorf("token: jwt.refresh.secretcode is still set to the placeholder value %q", token.PlaceholderSecret)
	}

	token.InitToken(issuer, access, refresh)
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
//
// Close is a no-op: the token package holds no resources that require release.
//
// Register this component after env/log and before gin, alongside auth/component
// if the service also uses Casbin authorization:
//
//	Add(envComp.Component("config/app.toml")).
//	Add(logComp.Component()).
//	AddParallel(dbComp.Component()).
//	PreReady(initServices).
//	Add(tokenComp.Component()).
//	Add(ginComp.Component(func(r *gin.Engine) { ... })).
func Component() bootstrap.IComponent {
	return bootstrap.Func("token", loadFromEnv, nil)
}
