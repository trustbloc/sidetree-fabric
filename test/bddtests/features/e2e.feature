#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@e2e
Feature:

  @e2e_s1
  Scenario: e2e
    Given the channel "mychannel" is created and all peers have joined
    And "test" chaincode "e2e_cc" is installed from path "github.com/trustbloc/e2e_cc" to all peers
    And "test" chaincode "e2e_cc" is instantiated from path "github.com/trustbloc/e2e_cc" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And chaincode "e2e_cc" is warmed up on all peers on the "mychannel" channel


    # Test transactions
    When client invokes chaincode "e2e_cc" with args "put,k1,20" on the "mychannel" channel
    And we wait 2 seconds
    And client queries chaincode "e2e_cc" with args "get,k1" on a single peer in the "peerorg1" org on the "mychannel" channel
    Then response from "e2e_cc" to client equal value "20"

    # sidetreetxn_cc
    And "system" chaincode "sidetreetxn_cc" is installed from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" to all peers
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" on the "mychannel" channel with args "" with endorsement policy "" with collection policy ""
    And chaincode "sidetreetxn_cc" is warmed up on all peers on the "mychannel" channel
