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

    Given variable "did_r" is assigned the value "TOKEN_DID_R"
    And variable "did_w" is assigned the value "TOKEN_DID_W"

    Given the channel "mychannel" is created and all peers have joined
    And the channel "yourchannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""
    And chaincode "sidetreetxn", version "v1", package ID "sidetreetxn:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy "dcas-cfg"
    And chaincode "document", version "v1", package ID "document:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" and collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "yourchannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""
    And chaincode "sidetreetxn", version "v1", package ID "sidetreetxn:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "yourchannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy "dcas-cfg"
    And chaincode "document", version "v1", package ID "document:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "yourchannel" channel with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" and collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "org1-mychannel-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"
    And fabric-cli context "org2-mychannel-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com,peer2.org2.example.com" and user "User1"
    And fabric-cli context "org1-yourchannel-context" is defined on channel "yourchannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"
    And fabric-cli context "org2-yourchannel-context" is defined on channel "yourchannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com,peer2.org2.example.com" and user "User1"

    And we wait 15 seconds

    # Configure the following Sidetree namespaces on channel 'mychannel'
    # - did:bloc:sidetree       - Path: /sidetree/0.0.1/identifiers, /sidetree/0.0.1/operations
    # - did:bloc:trustbloc.dev  - Path: /trustbloc.dev/identifiers, /trustbloc.dev/operations
    Then fabric-cli context "org1-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    Then fabric-cli context "org2-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

    # Configure the following Sidetree namespaces on channel 'yourchannel':
    # - did:bloc:yourdomain.com - Path: /yourdomain.com, /yourdomain.com/operations
    Then fabric-cli context "org1-yourchannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org1-config.json --noprompt"
    Then fabric-cli context "org2-yourchannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/yourchannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on yourchannel
    And we wait 10 seconds

  @create_did_doc
  Scenario: create valid did doc
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

    When client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/sidetree/0.0.1/identifiers" to resolve DID document with initial state
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/sidetree/0.0.1/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"
    When client sends request to "https://localhost:48327/sidetree/0.0.1/identifiers" to resolve DID document with alias "did:domain.com"
    Then check success response contains "#didDocumentHash"
    Then check success response contains "did:domain.com"

    When client sends request to "https://localhost:48426/trustbloc.dev/operations" to create DID document in namespace "did:bloc:trustbloc.dev"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/trustbloc.dev/identifiers" to resolve DID document with initial state
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/trustbloc.dev/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48426/yourdomain.com/operations" to create DID document in namespace "did:bloc:yourdomain.com"
    Then check success response contains "#didDocumentHash"

    When client sends request to "https://localhost:48327/yourdomain.com/identifiers" to resolve DID document with initial state
    Then check success response contains "#didDocumentHash"

    And we wait 10 seconds

    When client sends request to "https://localhost:48327/yourdomain.com/identifiers" to resolve DID document
    Then check success response contains "#didDocumentHash"

  @create_deactivate_did_doc
  Scenario: create and deactivate valid did doc
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

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
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

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
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

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
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"

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

  @sidetree_unauthorized
  Scenario: Attempt to access Sidetree endpoints without providing an auth token
    When an HTTP POST is sent to "https://localhost:48327/sidetree/0.0.1/operations" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48327/sidetree/0.0.1/identifiers/did:sidetree:1234" and the returned status code is 401

    # The following endpoints were configured with no authorization so they should be OK to access
    When an HTTP POST is sent to "https://localhost:48327/trustbloc.dev/operations" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 400
    When an HTTP GET is sent to "https://localhost:48327/trustbloc.dev/identifiers/did:bloc:trustbloc.dev:1234" and the returned status code is 404

    # Now provide valid auth tokens
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"
    When an HTTP POST is sent to "https://localhost:48327/sidetree/0.0.1/operations" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 400
    When an HTTP GET is sent to "https://localhost:48327/sidetree/0.0.1/identifiers/did:sidetree:1234" and the returned status code is 404

  @version_and_protocol_params
  Scenario: Version and protocol parameters
    # Protocol at time (block number) 50
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/version?time=50"
    And the JSON path "version" of the response equals "0.1.1"
    And the JSON path "genesisTime" of the numeric response equals "20"
    And the JSON path "hashAlgorithm" of the numeric response equals "5"
    And the JSON path "multihashAlgorithm" of the numeric response equals "18"
    And the JSON path "maxOperationCount" of the numeric response equals "30"
    And the JSON path "maxOperationSize" of the numeric response equals "200000"
    And the JSON path "maxAnchorFileSize" of the numeric response equals "1000000"
    And the JSON path "maxProofFileSize" of the numeric response equals "1000000"
    And the JSON path "maxMapFileSize" of the numeric response equals "1000000"
    And the JSON path "maxChunkFileSize" of the numeric response equals "10000000"
    And the JSON path "compressionAlgorithm" of the response equals "GZIP"
    And the JSON path "patches" of the array response is not empty
    And the JSON path "signatureAlgorithms" of the array response is not empty
    And the JSON path "keyAlgorithms" of the array response is not empty

    # Protocol at time (block number) 2000
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/version?time=2000"
    And the JSON path "version" of the response equals "0.1.2"
    And the JSON path "genesisTime" of the numeric response equals "1000"
    And the JSON path "hashAlgorithm" of the numeric response equals "5"
    And the JSON path "multihashAlgorithm" of the numeric response equals "18"
    And the JSON path "maxOperationCount" of the numeric response equals "50"
    And the JSON path "maxOperationSize" of the numeric response equals "300000"
    And the JSON path "maxAnchorFileSize" of the numeric response equals "2000000"
    And the JSON path "maxProofFileSize" of the numeric response equals "2000000"
    And the JSON path "maxMapFileSize" of the numeric response equals "2000000"
    And the JSON path "maxChunkFileSize" of the numeric response equals "20000000"
    And the JSON path "compressionAlgorithm" of the response equals "GZIP"
    And the JSON path "patches" of the array response is not empty
    And the JSON path "signatureAlgorithms" of the array response is not empty
    And the JSON path "keyAlgorithms" of the array response is not empty

    # Current protocol
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/version"
    # We can't check for actual version because we don't know how many blocks have been created
    # by the tests so far so we don't know which protocol is current
    And the JSON path "version" of the response is not empty
    And the JSON path "hashAlgorithm" of the numeric response equals "5"
    And the JSON path "multihashAlgorithm" of the numeric response equals "18"
    And the JSON path "compressionAlgorithm" of the response equals "GZIP"
    And the JSON path "signatureAlgorithms" of the array response is not empty
    And the JSON path "keyAlgorithms" of the array response is not empty
