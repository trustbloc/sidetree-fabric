// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests

require (
	github.com/cucumber/godog v0.8.1
	github.com/hyperledger/fabric-protos-go v0.0.0-20191121202242-f5500d5e3e85
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/viper v1.3.2
	github.com/trustbloc/fabric-peer-test-common v0.1.2-0.20200213155832-06af5163b73f
	github.com/trustbloc/sidetree-core-go v0.1.2-0.20200220212110-5676c1827fc3
)

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.2-0.20200205170340-c69bba6d7b81

replace github.com/hyperledger/fabric-sdk-go => github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20200203184105-5f7f0b025d89

go 1.13
