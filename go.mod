// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/loads v0.19.0
	github.com/go-openapi/runtime v0.19.0
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190429134815-48bb0d199e2c
	github.com/hyperledger/fabric/extensions v0.0.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.3.0
	github.com/trustbloc/fabric-peer-ext v0.0.0
	github.com/trustbloc/sidetree-core-go v0.0.0-20190604193932-b3a21a189580
	github.com/trustbloc/sidetree-node v0.0.0-20190605161025-0df7c418272b
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6 // indirect
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.0.0-20190617203614-83c67efef785

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20190617220514-28e5de45210b

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.0.0-20190617220514-28e5de45210b
