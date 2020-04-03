#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@did-sidetree
Feature:
  Background: Setup
    Given DCAS collection config "dcas-cfg" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=
    Given off-ledger collection config "diddoc-cfg" is defined for collection "diddoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=1, and timeToLive=
    Given off-ledger collection config "fileidx-cfg" is defined for collection "fileidxdoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=1, and timeToLive=
    Given off-ledger collection config "meta-data-cfg" is defined for collection "meta_data" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=1, and timeToLive=

    Given the channel "mychannel" is created and all peers have joined
    And the channel "yourchannel" is created and all peers have joined

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-cfg"
    And "system" chaincode "document" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-cfg"
    And "system" chaincode "document" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"
    And fabric-cli context "yourchannel" is defined on channel "yourchannel" with org "peerorg2", peers "peer0.org2.example.com,peer2.org1.example.com" and user "User1"

    And we wait 15 seconds

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
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/document" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48426/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48526/trustbloc.dev" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:bloc:trustbloc.dev"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/trustbloc.dev" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48426/trustbloc.dev" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48526/yourdomain.com" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:bloc:yourdomain.com"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/yourdomain.com" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48426/yourdomain.com" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @batch_writer_recovery
  Scenario: Batch writer recovers from peers down
    # Stop all of the peers in org2 so that processing of the batch fails (since we need two orgs for endorsement).
    Given container "peer0.org2.example.com" is stopped
    And container "peer1.org2.example.com" is stopped
    And we wait 2 seconds

    # Send the operation to peer0.org1.
    When client sends request to "https://localhost:48326/document" to create DID document "fixtures/testdata/didDocument2.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    # Stop peer0.org1 after sending it an operation. The operation should have
    # been saved to a persistent queue so that when it comes up it will be able to process it.
    Then container "peer0.org1.example.com" is stopped
    Then container "peer0.org1.example.com" is started

    # Upon starting up, peer0.org1 will try to process the operation but will fail since all peers in org2 are down.
    Then we wait 30 seconds

    Given container "peer0.org2.example.com" is started
    And container "peer1.org2.example.com" is started

    # Wait for the peers to come up and the batch writer to cut the batch
    And we wait 30 seconds

    # Retrieve the document from another peer since, by this time, the operation should have
    # been processed and distributed to all peers.
    When client sends request to "https://localhost:48626/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @invalid_config_update
  Scenario: Invalid configuration
    Given fabric-cli context "mychannel" is used
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-protocol-config.json --noprompt" then the error response should contain "algorithm not supported"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-config.json --noprompt" then the error response should contain "field 'BatchWriterTimeout' must contain a value greater than 0"
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-sidetree-peer-config.json --noprompt" then the error response should contain "field 'BasePath' must begin with '/'"

  @create_revoke_did_doc
  Scenario: create and revoke valid did doc
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "#didDocumentHash"
    When client sends request to "https://localhost:48526/document" to revoke DID document
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check error response contains "document is no longer available"

  @create_recover_did_doc
  Scenario: create and recover did doc
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48526/document" to recover DID document "fixtures/testdata/recover.json"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "recoveryKey"

  @create_update_did_doc
  Scenario: create and update valid did doc
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "#didDocumentHash"
    When client sends request to "https://localhost:48526/document" to update DID document path "/publicKey/0/type" with value "updatedValue"
    Then we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "updatedValue"


    @create_add_remove_public_key
    Scenario: add and remove public keys
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48526/document" to add public key with ID "newKey" to DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "newKey"

    When client sends request to "https://localhost:48526/document" to remove public key with ID "newKey" from DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response does NOT contain "newKey"

    @create_add_remove_services
    Scenario: add and remove service endpoints
    When client sends request to "https://localhost:48526/document" to create DID document "fixtures/testdata/didDocument.json" in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48526/document" to add service endpoint with ID "newService" to DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response contains "newService"

    When client sends request to "https://localhost:48526/document" to remove service endpoint with ID "newService" from DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48526/document" to resolve DID document
    Then check success response does NOT contain "newService"