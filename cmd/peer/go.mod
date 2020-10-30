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

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.1.5-0.20201015201411-75a48d16a707

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-peer-ext/mod/peer v0.0.0-20201016134740-812f1a67d3f5

replace github.com/trustbloc/fabric-peer-ext => github.com/trustbloc/fabric-peer-ext v0.1.5-0.20201030154508-f88b1c62fe4d

replace github.com/trustbloc/sidetree-fabric => ../..

replace github.com/spf13/viper2015 => github.com/spf13/viper v0.0.0-20150908122457-1967d93db724

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.5-0.20201005203042-9fe8149374fc

go 1.13
