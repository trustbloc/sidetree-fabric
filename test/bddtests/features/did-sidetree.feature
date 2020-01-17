#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@did-sidetree
Feature:
  Background: Setup
    Given DCAS collection config "dcas-mychannel" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given DCAS collection config "docs-mychannel" is defined for collection "docs" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given off-ledger collection config "meta_data_coll" is defined for collection "meta_data" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=60m

    Given the channel "mychannel" is created and all peers have joined

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-mychannel"
    And "system" chaincode "document_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "docs-mychannel,meta_data_coll"

    And we wait 10 seconds

    Given variable "org1Config" is assigned config from file "./fixtures/config/fabric/org1-config.json"
    And variable "org2Config" is assigned config from file "./fixtures/config/fabric/org2-config.json"

    When client invokes chaincode "configscc" with args "save,${org1Config}" on the "mychannel" channel
    And client invokes chaincode "configscc" with args "save,${org2Config}" on the "mychannel" channel
    And we wait 3 seconds

  @create_did_doc
  Scenario: create valid did doc
    When client sends request to "http://localhost:48426/document" to create DID document "fixtures/testdata/didDocument.json" as "queryParameter"
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" as "JSON"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "http://localhost:48626/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @did-sidetree-batch-writer-recovery
  Scenario: Batch writer recovers from peers down
    Given container "peer0.org2.example.com" is stopped
    And container "peer1.org2.example.com" is stopped
    And we wait 2 seconds

    When client sends request to "http://localhost:48326/document" to create DID document "fixtures/testdata/didDocument2.json" as "JSON"
    Then check success response contains "#didDocumentHash"

    Then we wait 10 seconds

    Given container "peer0.org2.example.com" is started
    And container "peer1.org2.example.com" is started

    # Wait for the peers to come up and the batch writer to cut the batch
    And we wait 30 seconds

    When client sends request to "http://localhost:48626/document" to resolve DID document
    Then check success response contains "#didDocumentHash"
