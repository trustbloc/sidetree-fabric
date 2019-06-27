#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@did-sidetree
Feature:

  @create_did_doc
  Scenario: create valid did doc

    Given DCAS collection config "dcas-mychannel" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=6000s
    Given DCAS collection config "docs-mychannel" is defined for collection "docs" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=3, maxPeerCount=3, and timeToLive=6000s

    Given the channel "mychannel" is created and all peers have joined

    And "system" chaincode "sidetreetxn_cc" is installed from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" to all peers
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" on the "mychannel" channel with args "" with endorsement policy "" with collection policy "dcas-mychannel"
    And chaincode "sidetreetxn_cc" is warmed up on all peers on the "mychannel" channel

    And "system" chaincode "document_cc" is installed from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc" to all peers
    And "system" chaincode "document_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc" on the "mychannel" channel with args "" with endorsement policy "" with collection policy "docs-mychannel"
    And chaincode "document_cc" is warmed up on all peers on the "mychannel" channel

    When client sends request to create DID document "fixtures/testdata/didDocument.json" as "queryParameter"
    Then check success response contains "#didDocumentHash"

    When client sends request to create DID document "fixtures/testdata/didDocument.json" as "JSON"
    Then check success response contains "#didDocumentHash"

    # batch writer needs some time to cut batch
    Then we wait 5 seconds

    When client sends request to resolve DID document
    Then check success response contains "#didDocumentHash"
