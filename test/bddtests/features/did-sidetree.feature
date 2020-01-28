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
    And the channel "yourchannel" is created and all peers have joined

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-mychannel"
    And "system" chaincode "document_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "docs-mychannel,meta_data_coll"

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-mychannel"
    And "system" chaincode "document_cc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "docs-mychannel,meta_data_coll"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"
    And fabric-cli context "yourchannel" is defined on channel "yourchannel" with org "peerorg2", peers "peer0.org2.example.com,peer2.org1.example.com" and user "User1"

    And we wait 10 seconds

    # Configure the following Sidetree namespaces on channel 'mychannel'
    # - did:bloc:sidetree       - Path: /document
    # - did:bloc:trustbloc.dev  - Path: /trustbloc.dev
    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

    # Configure the following Sidetree namespaces on channel 'yourchannel':
    # - did:bloc:yourdomain.com - Path: /yourdomain.com
    Then fabric-cli context "yourchannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on yourchannel
    And we wait 10 seconds

  @create_did_doc
  Scenario: create valid did doc
    When client sends request to "http://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48426/document" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "http://localhost:48426/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48526/trustbloc.dev" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:bloc:trustbloc.dev"
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48426/trustbloc.dev" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "http://localhost:48426/trustbloc.dev" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48526/yourdomain.com" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:bloc:yourdomain.com"
    Then check success response contains "#didDocumentHash"

    When client sends request to "http://localhost:48426/yourdomain.com" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "http://localhost:48426/yourdomain.com" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @did-sidetree-batch-writer-recovery
  Scenario: Batch writer recovers from peers down
    Given container "peer0.org2.example.com" is stopped
    And container "peer1.org2.example.com" is stopped
    And we wait 2 seconds

    When client sends request to "http://localhost:48326/document" to create DID document "fixtures/testdata/didDocument2.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    Then we wait 10 seconds

    Given container "peer0.org2.example.com" is started
    And container "peer1.org2.example.com" is started

    # Wait for the peers to come up and the batch writer to cut the batch
    And we wait 30 seconds

    When client sends request to "http://localhost:48626/document" to resolve DID document
    Then check success response contains "#didDocumentHash"
