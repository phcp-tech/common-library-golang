package loader

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

const (
	waitHttpServerStartupTimeout = 100 * time.Millisecond
)

func LoadFromEnv(router *gin.Engine) (httpserver.Runner, error) {
	// create runner synchronously so returned runner is non-nil
	if strings.EqualFold(env.Env().String("app.runmode"), "aws_lambda") {
		slog.Info("Http server is running under AWS-LAMBDA.")
		return lambda.NewHttpServer(), nil
	}

	// create a http runner.
	port := env.Env().String("http.server.port")
	slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s.", port))
	runner := httpserver.NewHttpServer(httpserver.Config{Port: port})

	// serverErr channel is used to capture errors from the server goroutine, including panics and startup errors. Not Exit in this goroutine.
	serverErr := make(chan error, 1)

	// start the server asynchronously; pass runner as param to avoid closure races
	go func(run httpserver.Runner, r *gin.Engine) {
		// only can recover panics from Start, can not recover errors returned by Start
		defer func() {
			if rec := recover(); rec != nil {
				serverErr <- fmt.Errorf("panic in http server goroutine: %v\nstack:%s", rec, string(debug.Stack()))
			}
		}()
		serverErr <- run.Start(r)
	}(runner, router)

	// wait briefly for an immediate startup error to avoid a race where
	// the goroutine hasn't written back an error yet. If no error is
	// received within the timeout we assume startup succeeded.
	select {
	case err := <-serverErr:
		return runner, err
	case <-time.After(waitHttpServerStartupTimeout):
		// no immediate error; continue
	}

	return runner, nil
}
