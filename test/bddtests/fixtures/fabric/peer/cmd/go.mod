// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/sidetree-fabric/test/bddtests/fixtures/fabric/peer/cmd

require (
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/spf13/viper v0.0.0-20150908122457-1967d93db724
)

replace github.com/hyperledger/fabric => github.com/trustbloc/fabric-mod v0.0.0-20190605152521-6547615cb978

replace github.com/hyperledger/fabric/extensions => github.com/trustbloc/fabric-mod/extensions v0.0.0-20190605152521-6547615cb978
