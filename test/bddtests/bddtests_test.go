/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/spf13/viper"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

var context *bddtests.BDDContext

func TestMain(m *testing.M) {
	projectPath, err := filepath.Abs("../..")
	if err != nil {
		panic(err.Error())
	}
	if err := os.Setenv("PROJECT_PATH", projectPath); err != nil {
		panic(err.Error())
	}

	// default is to run all tests with tag @all
	tags := "all"

	flag.Parse()
	cmdTags := flag.CommandLine.Lookup("test.run")
	if cmdTags != nil && cmdTags.Value != nil && cmdTags.Value.String() != "" {
		tags = cmdTags.Value.String()
	}

	initBDDConfig()

	compose := os.Getenv("DISABLE_COMPOSITION") != "true"
	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		s.BeforeSuite(func() {
			if compose {
				if err := context.Composition().Up(); err != nil {
					panic(fmt.Sprintf("Error composing system in BDD context: %s", err))
				}
			}
		})

		s.AfterSuite(func() {
			if compose {
				composition := context.Composition()
				if err := composition.GenerateLogs(); err != nil {
					logger.Warnf("Error generating logs: %s", err)
				}
				if _, err := composition.Decompose(); err != nil {
					logger.Warnf("Error decomposing: %s", err)
				}
			}
		})

		FeatureContext(s)
	}, godog.Options{
		Tags:          tags,
		Format:        "progress",
		Paths:         []string{"features"},
		Randomize:     time.Now().UTC().UnixNano(), // randomize scenario execution order
		Strict:        true,
		StopOnFailure: true,
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	peersMspID := make(map[string]string)
	peersMspID["peer0.org1.example.com"] = "Org1MSP"
	peersMspID["peer1.org1.example.com"] = "Org1MSP"
	peersMspID["peer0.org2.example.com"] = "Org2MSP"
	peersMspID["peer1.org2.example.com"] = "Org2MSP"

	var err error
	context, err = bddtests.NewBDDContext([]string{"peerorg1", "peerorg2"}, "orderer.example.com", "./fixtures/config/sdk-client/",
		"config.yaml", peersMspID, "../../.build/cc", "../../.build/cc")
	if err != nil {
		panic(fmt.Sprintf("Error returned from NewBDDContext: %s", err))
	}

	composeProjectName := strings.Replace(GenerateUUID(), "-", "", -1)
	composition, err := bddtests.NewDockerCompose(composeProjectName, "docker-compose.yml", "./fixtures")
	if err != nil {
		panic(fmt.Sprintf("Error creating a Docker-Compose client: %s", err))
	}
	context.SetComposition(composition)

	// Context is shared between tests - for now
	// Note: Each test after NewcommonSteps. should add unique steps only
	bddtests.NewCommonSteps(context).RegisterSteps(s)
	bddtests.NewDockerSteps(context).RegisterSteps(s)
	NewOffLedgerSteps(context).RegisterSteps(s)
	NewSidetreeSteps(context).RegisterSteps(s)
	NewDIDSideSteps(context).RegisterSteps(s)
	NewFabricCLISteps(context).RegisterSteps(s)
	NewFileHandlerSteps(context).RegisterSteps(s)
}

func initBDDConfig() {
	replacer := strings.NewReplacer(".", "_")

	viper.AddConfigPath("./fixtures/config/sdk-client/")
	viper.SetConfigName("config")
	viper.SetEnvPrefix("core")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error reading config file: %s \n", err)
		os.Exit(1)
	}
}
