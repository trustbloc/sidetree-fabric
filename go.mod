// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric

require (
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/bluele/gcache v0.0.0-20190301044115-79ae3b2d8680
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible
	github.com/golang/protobuf v1.3.3
	github.com/gorilla/mux v1.7.3
	github.com/hyperledger/fabric v2.0.0+incompatible
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200128192331-2d899240a7ed
	github.com/hyperledger/fabric-config v0.0.5
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta2.0.20200730104005-bdbc9bd37260
	github.com/hyperledger/fabric/extensions v0.0.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/rs/cors v1.7.0
	github.com/spf13/viper v1.7.0
	github.com/spf13/viper2015 v1.3.2
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.1-0.20190625010220-02440ea7a285
	github.com/trustbloc/fabric-peer-ext v0.0.0
	github.com/trustbloc/sidetree-core-go v0.1.4-0.20200731163924-244e4e845359
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.1.4-0.20200731181404-3659bd5007fe

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20200804135952-4b9196b6678d

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.1.4-0.20200804135952-4b9196b6678d

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.4-0.20200626180529-18936b36feca

replace github.com/spf13/viper2015 => github.com/spf13/viper v0.0.0-20150908122457-1967d93db724

go 1.13
