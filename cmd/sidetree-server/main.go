/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-node/pkg/requesthandler"
	"github.com/trustbloc/sidetree-node/restapi"
	"github.com/trustbloc/sidetree-node/restapi/operations"
)

const didDocNamespace = "did:sidetree:"

func main() {

	swaggerSpec, handlerErr := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if handlerErr != nil {
		panic(handlerErr)
	}

	api := operations.NewSidetreeAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer serverShutdown(server)

	parser := flags.NewParser(server, flags.Default)
	// Custom configure flags
	configureFlags()
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			panic(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	// Custom: Configure handler
	handler, handlerErr := configureAPI(api)
	if handlerErr != nil {
		panic(handlerErr)
	}

	server.SetHandler(handler)

	if err := server.Serve(); err != nil {
		panic(err)
	}

}

func serverShutdown(server *restapi.Server) {
	if err := server.Shutdown(); err != nil {
		log.Println(fmt.Printf("error during server shutdown: %s", err.Error()))
	}

	log.Println("shutdown sidetree node...")

}

func configureFlags() {
	// Set command line options from environment variables if available
	args := []string{
		"scheme",
		"cleanup-timeout",
		"graceful-timeout",
		"max-header-size",
		"socket-path",
		"host",
		"port",
		"listen-limit",
		"keep-alive",
		"read-timeout",
		"write-timeout",
		"tls-host",
		"tls-port",
		"tls-certificate",
		"tls-key",
		"tls-ca",
		"tls-listen-limit",
		"tls-keep-alive",
		"tls-read-timeout",
		"tls-write-timeout",
	}
	for _, a := range args {
		if envVar := os.Getenv(fmt.Sprintf("SIDETREE_NODE_%s", strings.Replace(strings.ToUpper(a), "-", "_", -1))); envVar != "" {
			os.Args = append(os.Args, fmt.Sprintf("--%s=%s", a, envVar))
		}
	}
}

func configureAPI(api *operations.SidetreeAPI) (http.Handler, error) {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()
	api.ApplicationJoseProducer = runtime.JSONProducer()
	api.JSONProducer = runtime.JSONProducer()

	var logger = logrus.New()
	var config = viper.New()

	config.SetEnvPrefix("SIDETREE_NODE")
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	logger.Info("starting sidetree node...")

	channelProvider, err := context.GetChannelProvider(config)
	if err != nil {
		logger.Errorf("Failed to create channel provider: %s", err.Error())
		return nil, err
	}

	ctx, err := context.New(config, channelProvider)
	if err != nil {
		logger.Errorf("Failed to create new batch writer context: %s", err.Error())
		return nil, err
	}

	didDocStore := store.New(channelProvider, didDocNamespace)

	// create new batch writer
	// TODO: Make batch timeout configurable
	batchWriter, err := batch.New(ctx, batch.WithBatchTimeout(1*time.Second))
	if err != nil {
		logger.Errorf("Failed to create batch writer: %s", err.Error())
		return nil, err
	}

	// start routine for creating batches
	batchWriter.Start()

	// did document handler with did document validator for didDocNamespace
	didDocHandler := dochandler.New(
		didDocNamespace,
		ctx.Protocol(),
		didvalidator.New(didDocStore),
		batchWriter,
		processor.New(didDocStore),
	)

	didResolutionHandler := requesthandler.NewResolutionHandler(didDocNamespace, ctx.Protocol(), didDocHandler)
	didOperationHandler := requesthandler.NewOperationHandler(didDocNamespace, ctx.Protocol(), didDocHandler)

	api.PostDocumentHandler = operations.PostDocumentHandlerFunc(
		func(params operations.PostDocumentParams) middleware.Responder {
			return didOperationHandler.HandleOperationRequest(params.Request)
		},
	)
	api.GetDocumentDidOrDidDocumentHandler = operations.GetDocumentDidOrDidDocumentHandlerFunc(
		func(params operations.GetDocumentDidOrDidDocumentParams) middleware.Responder {
			return didResolutionHandler.HandleResolveRequest(params.DidOrDidDocument)
		},
	)
	api.ServerShutdown = func() {}

	return setupAndServe(api), nil
}

func setupAndServe(api *operations.SidetreeAPI) http.Handler {
	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
