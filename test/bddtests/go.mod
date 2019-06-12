// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20181230093703-5ccba6eab8d6 // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190429134815-48bb0d199e2c
	github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric v0.0.0-20190429134815-48bb0d199e2c
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/viper v1.0.2
	github.com/sykesm/zap-logfmt v0.0.2 // indirect
	github.com/trustbloc/fabric-peer-test-common v0.0.0-20190528215613-a7959c5ba3e1
	go.uber.org/zap v1.10.0 // indirect

)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric => github.com/trustbloc/fabric-sdk-go-ext/fabric v0.0.0-20190528182243-b95c24511993
