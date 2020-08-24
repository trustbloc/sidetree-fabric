// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/cmd/peer

require (
	github.com/hyperledger/fabric v2.0.0+incompatible
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper2015 v1.3.2
	github.com/trustbloc/sidetree-fabric v0.0.0
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.1.4-0.20200824170048-acc0cead7eb9

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20200824174524-030db816e7d6

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.1.4-0.20200824174524-030db816e7d6

replace github.com/trustbloc/sidetree-fabric => ../..

replace github.com/spf13/viper2015 => github.com/spf13/viper v0.0.0-20150908122457-1967d93db724

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.4-0.20200626180529-18936b36feca

go 1.13
