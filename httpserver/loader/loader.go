package loader

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

func LoadDefault(router *gin.Engine) (httpserver.Runner, error) {
	var runner httpserver.Runner
	// create runner synchronously so returned runner is non-nil
	if strings.EqualFold(env.Env().String("app.runmode"), "aws_lambda") {
		slog.Info("Http server is running under AWS-LAMBDA.")
		runner = lambda.NewHttpServer()
	} else {
		port := env.Env().String("http.server.port")
		slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s.", port))
		runner = httpserver.NewHttpServer(httpserver.Config{Port: port})
	}

	// start the server asynchronously; pass runner as param to avoid closure races
	go func(run httpserver.Runner, r *gin.Engine) {
		// only can recover panics from Start, can not recover errors returned by Start
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error(fmt.Sprintf("panic in http server goroutine: %v\n%s", rec, string(debug.Stack())))
				return
			}
		}()

		if err := run.Start(r); err != nil {
			slog.Error(fmt.Sprintf("Startup http server failed: %s.", err.Error()))
			os.Exit(1) // exit if http server fails to start
		}
	}(runner, router)

	return runner, nil
}
