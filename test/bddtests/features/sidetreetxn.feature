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

    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-cfg"
    And "system" chaincode "document" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"

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
    When client sends request to "https://localhost:48326/sidetree/0.1.3/sidetree/operations" to create DID document in namespace "did:sidetree"
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
    When client sends request to "https://localhost:48427/sidetree/0.1.3/sidetree" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @invalid_config_update
  Scenario: Invalid configuration
    Given fabric-cli context "mychannel" is used
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-protocol-config.json --noprompt" then the error response should contain "algorithm not supported"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-config.json --noprompt" then the error response should contain "field 'BatchWriterTimeout' must contain a value greater than 0"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-peer-config.json --noprompt" then the error response should contain "field 'BasePath' must begin with '/'"
