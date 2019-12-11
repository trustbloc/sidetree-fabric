// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/hyperledger/fabric-protos-go v0.0.0-20190823190507-26c33c998676
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/viper v1.3.2
	github.com/trustbloc/fabric-peer-test-common v0.0.0-20191029204148-075911ef5e01
	github.com/trustbloc/sidetree-core-go v0.1.0
)

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.1-0.20191126151100-5a61374c2e1b

go 1.13
