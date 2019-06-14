// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/loads v0.19.0
	github.com/go-openapi/runtime v0.19.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/hyperledger/fabric-amcl v0.0.0-20181230093703-5ccba6eab8d6 // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190328182020-93c3fcb272be
	github.com/jessevdk/go-flags v1.4.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.3.0
	github.com/sykesm/zap-logfmt v0.0.2 // indirect
	github.com/trustbloc/sidetree-core-go v0.0.0-20190604193932-b3a21a189580
	github.com/trustbloc/sidetree-node v0.0.0-20190605161025-0df7c418272b
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos => github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos v0.0.0-20190328182020-93c3fcb272be
