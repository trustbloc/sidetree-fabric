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
    Given off-ledger collection config "diddoc-cfg" is defined for collection "diddoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "fileidx-cfg" is defined for collection "fileidxdoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "meta-data-cfg" is defined for collection "meta_data" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=

    Given the channel "mychannel" is created and all peers have joined
    And the channel "yourchannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-cfg"
    And "system" chaincode "document" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-cfg"
    And "system" chaincode "document" is instantiated from path "in-process" on the "yourchannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"
    And fabric-cli context "yourchannel" is defined on channel "yourchannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com,peer2.org2.example.com" and user "User1"

    And we wait 15 seconds

    # Configure the following Sidetree namespaces on channel 'mychannel'
    # - did:bloc:sidetree       - Path: /sidetree/0.0.1/identifiers, /sidetree/0.0.1/operations
    # - did:bloc:trustbloc.dev  - Path: /trustbloc.dev/identifiers, /trustbloc.dev/operations
    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

    # Configure the following Sidetree namespaces on channel 'yourchannel':
    # - did:bloc:yourdomain.com - Path: /yourdomain.com, /yourdomain.com/operations
    Then fabric-cli context "yourchannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on yourchannel
    And we wait 10 seconds

  @create_did_doc
  Scenario: create valid did doc
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/sidetree/0.0.1/identifiers" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/trustbloc.dev/operations" to create DID document in namespace "did:bloc:trustbloc.dev"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/trustbloc.dev/identifiers" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/trustbloc.dev/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/yourdomain.com/operations" to create DID document in namespace "did:bloc:yourdomain.com"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/yourdomain.com/identifiers" to resolve DID document with initial value
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/yourdomain.com/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @create_deactivate_did_doc
  Scenario: create and deactivate valid did doc
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to deactivate DID document
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check error response contains "document is no longer available"

  @create_recover_did_doc
  Scenario: create and recover did doc
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to recover DID document
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "recoveryKey"

  @create_add_remove_public_key
  Scenario: add and remove public keys
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to add public key with ID "newKey" to DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "newKey"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to remove public key with ID "newKey" from DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response does NOT contain "newKey"

  @create_add_remove_services
  Scenario: add and remove service endpoints
    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"
    And we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to add service endpoint with ID "newService" to DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "newService"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to remove service endpoint with ID "newService" from DID document
    Then we wait 10 seconds

    When client sends request to "https://localhost:48426/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response does NOT contain "newService"