/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/spf13/viper"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

var composition *bddtests.Composition

func TestMain(m *testing.M) {

	// default is to run all tests with tag @all
	tags := "all"

	flag.Parse()
	cmdTags := flag.CommandLine.Lookup("test.run")
	if cmdTags != nil && cmdTags.Value != nil && cmdTags.Value.String() != "" {
		tags = cmdTags.Value.String()
	}

	initBDDConfig()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		s.BeforeSuite(func() {

			if os.Getenv("DISABLE_COMPOSITION") != "true" {

				// Need a unique name, but docker does not allow '-' in names
				composeProjectName := strings.Replace(GenerateUUID(), "-", "", -1)
				newComposition, err := bddtests.NewComposition(composeProjectName, "docker-compose.yml", "./fixtures")
				if err != nil {
					panic(fmt.Sprintf("Error composing system in BDD context: %s", err))
				}

				composition = newComposition

				fmt.Println("docker-compose up ... waiting for peer to start ...")
				testSleep := 5
				if os.Getenv("TEST_SLEEP") != "" {
					testSleep, _ = strconv.Atoi(os.Getenv("TEST_SLEEP"))
				}
				fmt.Printf("*** testSleep=%d", testSleep)
				time.Sleep(time.Second * time.Duration(testSleep))
			}

		})

		s.AfterSuite(func() {
			if composition != nil {
				composition.GenerateLogs("./fixtures")
				composition.Decompose("./fixtures")
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
	context, err := bddtests.NewBDDContext([]string{"peerorg1", "peerorg2"}, "orderer.example.com", "./fixtures/config/sdk-client/",
		"config.yaml", peersMspID, "../../.build/cc", "./fixtures/testdata")
	if err != nil {
		panic(fmt.Sprintf("Error returned from NewBDDContext: %s", err))
	}

	// Context is shared between tests - for now
	// Note: Each test after NewcommonSteps. should add unique steps only
	bddtests.NewCommonSteps(context).RegisterSteps(s)
	NewOffLedgerSteps(context).RegisterSteps(s)
	NewSidetreeSteps(context).RegisterSteps(s)
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
