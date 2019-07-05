// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/loads v0.19.0
	github.com/go-openapi/runtime v0.19.3
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190429134815-48bb0d199e2c
	github.com/hyperledger/fabric/extensions v0.0.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/trustbloc/fabric-peer-ext v0.0.0
	github.com/trustbloc/sidetree-core-go v0.0.0-20190703191711-b3562d2d9c99
	github.com/trustbloc/sidetree-node v0.0.0-20190627183933-1e09f18640f3
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6 // indirect
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.0.0-20190617203614-83c67efef785

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20190626183708-8f13fb5c70f7

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.0.0-20190626183708-8f13fb5c70f7
