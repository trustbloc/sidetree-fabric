/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	sidetreecontext "github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

const didDocNamespace = "did:sidetree:"

var logger = logrus.New()
var config = viper.New()

func main() {
	config.SetEnvPrefix("SIDETREE_NODE")
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	logger.Info("starting sidetree node...")

	channelProvider, err := sidetreecontext.GetChannelProvider(config)
	if err != nil {
		logger.Errorf("Failed to create channel provider: %s", err.Error())
		panic(err)
	}

	ctx, err := sidetreecontext.New(config, channelProvider)
	if err != nil {
		logger.Errorf("Failed to create new batch writer context: %s", err.Error())
		panic(err)
	}

	didDocStore := store.New(channelProvider, didDocNamespace)

	// create new batch writer
	// TODO: Make batch timeout configurable
	batchWriter, err := batch.New(ctx, batch.WithBatchTimeout(1*time.Second))
	if err != nil {
		logger.Errorf("Failed to create batch writer: %s", err.Error())
		panic(err)
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

	restSvc := httpserver.New(
		getListenURL(),
		diddochandler.NewUpdateHandler(didDocHandler),
		diddochandler.NewResolveHandler(didDocHandler),
	)

	if restSvc.Start() != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt
	<-interrupt

	// Shut down all services
	batchWriter.Stop()

	if err := restSvc.Stop(context.Background()); err != nil {
		logger.Errorf("Error stopping REST service: %s", err)
	}
}

func getListenURL() string {
	host := config.GetString("host")
	if host == "" {
		host = "0.0.0.0"
	}
	port := config.GetInt("port")
	if port == 0 {
		panic("port is not set")
	}
	return fmt.Sprintf("%s:%d", host, port)
}
