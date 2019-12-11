#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@sanity
Feature:

  @sanity_s1
  Scenario: sanity
    Given the channel "mychannel" is created and all peers have joined
    And "test" chaincode "e2e_cc" is installed from path "github.com/trustbloc/sidetree-fabric/test/chaincode/e2e_cc" to all peers
    And "test" chaincode "e2e_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/test/chaincode/e2e_cc" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And chaincode "e2e_cc" is warmed up on all peers on the "mychannel" channel

    # Test transactions
    When client invokes chaincode "e2e_cc" with args "put,k1,20" on the "mychannel" channel
    And we wait 2 seconds
    And client queries chaincode "e2e_cc" with args "get,k1" on a single peer in the "peerorg1" org on the "mychannel" channel
    Then response from "e2e_cc" to client equal value "20"
