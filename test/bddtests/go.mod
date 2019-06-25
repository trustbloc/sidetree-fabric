// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric v0.0.0-20190429134815-48bb0d199e2c
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/spf13/viper v1.0.2
	github.com/trustbloc/fabric-peer-test-common v0.0.0-20190528215613-a7959c5ba3e1
	golang.org/x/crypto v0.0.0-20190320223903-b7391e95e576 // indirect
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092 // indirect
	golang.org/x/sys v0.0.0-20190321052220-f7bb7a8bee54 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect

)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric => github.com/trustbloc/fabric-sdk-go-ext/fabric v0.0.0-20190528182243-b95c24511993
