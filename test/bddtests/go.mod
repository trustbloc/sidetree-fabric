// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/go-openapi/swag v0.19.0
	github.com/hyperledger/fabric-protos-go v0.0.0-20190823190507-26c33c998676
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/viper v1.3.2
	github.com/trustbloc/fabric-peer-test-common v0.0.0-20191001161824-e89c26cf9121
	github.com/trustbloc/sidetree-core-go v0.0.0-20191017153620-a489e16494b3
	github.com/trustbloc/sidetree-node v0.0.0-20191017194101-9674831a60aa
)

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.0.0-20191001172134-1815f5c382ff

go 1.13
