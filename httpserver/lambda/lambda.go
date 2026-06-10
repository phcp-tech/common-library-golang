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

// Package lambda implements httpserver.Runner for AWS Lambda deployments.
// Import this sub-package only in services that run on AWS Lambda — importing
// it pulls in the AWS Lambda SDK dependencies. Services that run as plain
// HTTP servers should use the parent httpserver package directly.

package lambda

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/phcp-tech/common-library-golang/httpserver"
)

// Compile-time check that lambdaRunner implements httpserver.Runner.
var _ httpserver.Runner = (*lambdaRunner)(nil)

// lambdaRunner implements httpserver.Runner for AWS Lambda.
// It bridges the Lambda APIGatewayProxyRequest event to a standard http.Handler
// via aws-lambda-go-api-proxy/httpadapter, so any http.Handler (including
// *gin.Engine) can be used without modification.
type lambdaRunner struct{}

// NewHttpServer returns an httpserver.Runner for AWS Lambda deployments.
func NewHttpServer() httpserver.Runner { return &lambdaRunner{} }

// Start registers handler with the Lambda runtime and blocks until the process
// is terminated by AWS. It always returns nil.
func (r *lambdaRunner) Start(handler http.Handler) error {
	adapter := httpadapter.New(handler)
	awslambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// If no name is provided in the HTTP request body, throw an error
		return adapter.ProxyWithContext(ctx, req)
	})
	return nil
}

// Shutdown is a no-op; the AWS runtime manages the Lambda lifecycle.
func (r *lambdaRunner) Shutdown(_ context.Context) error {
	return nil
}
