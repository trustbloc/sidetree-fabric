// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/bluele/gcache v0.0.0-20190301044115-79ae3b2d8680
	github.com/btcsuite/btcutil v0.0.0-20170419141449-a5ecb5d9547a
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/hyperledger/fabric-protos-go v0.0.0-20190823190507-26c33c998676 // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20190930220855-cea2ffaf627c
	github.com/hyperledger/fabric/extensions v0.0.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/trustbloc/fabric-peer-ext v0.0.0
	github.com/trustbloc/sidetree-core-go v0.0.0-20191028174235-80e3b92d0da1
	go.uber.org/atomic v1.4.0 // indirect
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.0.0-20190821180934-5941d21b98c6

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20191010163548-1ab7e444ef7c

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.0.0-20191010163548-1ab7e444ef7c

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.0.0-20191001172134-1815f5c382ff

go 1.13
