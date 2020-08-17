#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@sidetreetxn
Feature:
  Background: Setup
    Given DCAS collection config "dcas-cfg" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=
    Given off-ledger collection config "diddoc-cfg" is defined for collection "diddoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "fileidx-cfg" is defined for collection "fileidxdoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "meta-data-cfg" is defined for collection "meta_data" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=

    Given variable "did_r" is assigned the value "TOKEN_DID_R"
    And variable "did_w" is assigned the value "TOKEN_DID_W"
    Then the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""
    And chaincode "sidetreetxn", version "v1", package ID "sidetreetxn:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy "dcas-cfg"
    And chaincode "document", version "v1", package ID "document:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" and collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "org1-mychannel-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"
    And fabric-cli context "org2-mychannel-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com,peer2.org2.example.com" and user "User1"

    And we wait 10 seconds

    # Configure the following Sidetree namespaces on channel 'mychannel'
    # - did:bloc:sidetree       - Path: /document
    # - did:bloc:trustbloc.dev  - Path: /trustbloc.dev
    Then fabric-cli context "org1-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    Then fabric-cli context "org2-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @sidetree_1
  Scenario: Sidetree Txn Test
    # sidetree content test
    When client writes content "Hello World" using "sidetreetxn" and the "dcas" collection on the "mychannel" channel
    Then client verifies that written content at the returned address from "sidetreetxn" and the "dcas" collection matches original content on the "mychannel" channel

  @batch_writer_recovery
  Scenario: Batch writer recovers from peers down
    # Stop all of the peers in org2 so that processing of the batch fails (since we need two orgs for endorsement).
    Given container "peer0.org2.example.com" is stopped
    And container "peer1.org2.example.com" is stopped
    And container "peer2.org2.example.com" is stopped
    And we wait 2 seconds

    # Send the operation to peer0.org1.
    When client sends request to "https://localhost:48326/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    # Stop peer0.org1 after sending it an operation. The operation should have
    # been saved to a persistent queue so that when it comes up it will be able to process it.
    Then container "peer0.org1.example.com" is stopped
    Then container "peer0.org1.example.com" is started

    # Upon starting up, peer0.org1 will try to process the operation but will fail since all peers in org2 are down.
    Then we wait 30 seconds

    Given container "peer0.org2.example.com" is started
    And container "peer1.org2.example.com" is started
    And container "peer2.org2.example.com" is started

    # Wait for the peers to come up and the batch writer to cut the batch
    And we wait 30 seconds

    # Retrieve the document from another peer since, by this time, the operation should have
    # been processed and distributed to all peers.
    When client sends request to "https://localhost:48427/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @invalid_config_update
  Scenario: Invalid configuration
    Given fabric-cli context "org1-mychannel-context" is used
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-protocol-config.json --noprompt" then the error response should contain "algorithm not supported"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-config.json --noprompt" then the error response should contain "field 'BatchWriterTimeout' must contain a value greater than 0"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-peer-config.json --noprompt" then the error response should contain "field 'BasePath' must begin with '/'"

  @observer_failover
  Scenario: Active observer fails over to standby
    Given variable "peer0.org1" is assigned the value "https://localhost:48326/sidetree/0.0.1"
    And variable "peer1.org1" is assigned the value "https://localhost:48327/sidetree/0.0.1"
    And variable "peer2.org1" is assigned the value "https://localhost:48328/sidetree/0.0.1"
    And variable "peer0.org2" is assigned the value "https://localhost:48426/sidetree/0.0.1"
    And variable "peer1.org2" is assigned the value "https://localhost:48427/sidetree/0.0.1"
    And variable "peer2.org2" is assigned the value "https://localhost:48428/sidetree/0.0.1"

    # Write several Sidetree transactions. Scatter the requests across different endpoints to generate multiple
    # Sidetree transactions within the same block and across multiple blocks.
    When client sends request to "${peer0.org1}/operations,${peer2.org1}/operations,${peer0.org2}/operations,${peer2.org2}/operations" to create 500 DID documents using 10 concurrent requests

    # Take down the active observers in both orgs so that the standby observers can take over
    Given container "peer1.org1.example.com" is stopped
    And container "peer1.org2.example.com" is stopped
    Then we wait 60 seconds

    # Verify that all of the documents are there
    Then client sends request to "${peer0.org1}/identifiers,${peer2.org1}/identifiers,${peer0.org2}/identifiers,${peer2.org2}/identifiers" to verify the DID documents that were created

    Given container "peer1.org1.example.com" is started
    And container "peer1.org2.example.com" is started
    Then we wait 10 seconds

    # Write a few more Sidetree transactions to ensure everything's back to normal
    When client sends request to "${peer0.org1}/operations,${peer1.org1}/operations,${peer2.org1}/operations,${peer0.org2}/operations,${peer1.org2}/operations,${peer2.org2}/operations" to create 50 DID documents using 5 concurrent requests
    Then we wait 10 seconds

    # Verify that all of the documents are there
    Then client sends request to "${peer0.org1}/identifiers,${peer1.org1}/identifiers,${peer2.org1}/identifiers,${peer0.org2}/identifiers,${peer1.org2}/identifiers,${peer2.org2}/identifiers" to verify the DID documents that were created
