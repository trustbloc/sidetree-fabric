#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@file-handler
Feature:
  Background: Setup
    Given DCAS collection config "dcas-cfg" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=
    Given off-ledger collection config "diddoc-cfg" is defined for collection "diddoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "fileidx-cfg" is defined for collection "fileidxdoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "meta-data-cfg" is defined for collection "meta_data" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=

    Given variable "content_r" is assigned the value "TOKEN_CONTENT_R"
    And variable "content_w" is assigned the value "TOKEN_CONTENT_W"

    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""
    And chaincode "sidetreetxn", version "v1", package ID "sidetreetxn:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy "dcas-cfg"
    And chaincode "document", version "v1", package ID "document:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" and collection policy "diddoc-cfg,fileidx-cfg,meta-data-cfg"

    Given DCAS collection config "consortium-files-cfg" is defined for collection "consortium" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=
    And chaincode "file", version "v1", package ID "file:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" and collection policy "consortium-files-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "org1-mychannel-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com,peer2.org1.example.com" and user "User1"
    And fabric-cli context "org2-mychannel-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com,peer2.org2.example.com" and user "User1"

    And we wait 10 seconds

    Then fabric-cli context "org1-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    Then fabric-cli context "org2-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @upload_and_retrieve_files
  Scenario: upload files to DCAS and create index file on Sidetree that references them
    And the authorization bearer token for "POST" requests to path "/schema" is set to "${content_w}"
    And the authorization bearer token for "POST" requests to path "/.well-known/did-bloc" is set to "${content_w}"
    And the authorization bearer token for "GET" requests to path "/file" is set to "${content_r}"
    And the authorization bearer token for "POST" requests to path "/file" is set to "${content_w}"

    # Upload schema files
    When client sends request to "https://localhost:48326/schema" to upload file "fixtures/testdata/schemas/arrays.schema.json" with content type "application/json"
    Then the ID of the file is saved to variable "arraysSchemaID"
    # Create the schema file index Sidetree document
    Given variable "schemaIndexFile" is assigned the JSON value '{"fileIndex":{"basePath":"/schema","mappings":{"arrays.schema.json":"${arraysSchemaID}"}}}'
    When client sends request to "https://localhost:48426/file/operations" to create document with content "${schemaIndexFile}" in namespace "file:idx"
    Then the ID of the returned document is saved to variable "schemaIndexID"

    # Upload .well-known files
    When client sends request to "https://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/trustbloc.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownTrustblocID"
    When client sends request to "https://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/org1.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownOrg1ID"
    When client sends request to "https://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/org2.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownOrg2ID"
    # Create the .well-known file index Sidetree document
    Given variable "wellKnownIndexFile" is assigned the JSON value '{"fileIndex":{"basePath":"/.well-known/did-bloc","mappings":{"trustbloc.dev.json":"${wellKnownTrustblocID}","org1.dev.json":"${wellKnownOrg1ID}","org2.dev.json":"${wellKnownOrg2ID}"}}}'
    When client sends request to "https://localhost:48426/file/operations" to create document with content "${wellKnownIndexFile}" in namespace "file:idx"
    Then the ID of the returned document is saved to variable "wellKnownIndexID"

    # Update the ledger config to point to the index file documents
    Given variable "schemaHandlerConfig" is assigned the JSON value '{"BasePath":"/schema","ChaincodeName":"file","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"${schemaIndexID}"}'
    And variable "wellKnownHandlerConfig" is assigned the JSON value '{"BasePath":"/.well-known/did-bloc","ChaincodeName":"file","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"${wellKnownIndexID}"}'
    And variable "org1ConfigUpdate" is assigned the JSON value '{"MspID":"Org1MSP","Peers":[{"PeerID":"peer0.org1.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer1.org1.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer2.org1.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]}]}'
    And variable "org2ConfigUpdate" is assigned the JSON value '{"MspID":"Org2MSP","Peers":[{"PeerID":"peer0.org2.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer1.org2.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer2.org2.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]}]}'
    Then fabric-cli context "org1-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --config ${org1ConfigUpdate} --noprompt"
    Then fabric-cli context "org2-mychannel-context" is used
    And fabric-cli is executed with args "ledgerconfig update --config ${org2ConfigUpdate} --noprompt"

    # Resolve schema files
    When client sends request to "https://localhost:48326/schema/arrays.schema.json" to retrieve file
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"
    # geographical-location.schema.json should not be there until we upload it and update the index
    When client sends request to "https://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the response has status code 404 and error message "file not found"

    # Resolve .well-known files
    When client sends request to "https://localhost:48427/.well-known/did-bloc/trustbloc.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "trustbloc.dev"
    When client sends request to "https://localhost:48427/.well-known/did-bloc/org1.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "org1.dev"
    When client sends request to "https://localhost:48428/.well-known/did-bloc/org2.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "org2.dev"

    # Upload a new schema and update the schema index document
    When client sends request to "https://localhost:48326/schema" to upload file "fixtures/testdata/schemas/geographical-location.schema.json" with content type "application/json"
    Then the ID of the file is saved to variable "locationsSchemaID"
    # Update the schema file index Sidetree document
    Given variable "schemaPatch" is assigned the JSON patch '[{"op": "add", "path": "/fileIndex/mappings/geographical-location.schema.json", "value": "${locationsSchemaID}"}]'
    When client sends request to "https://localhost:48326/file/operations" to update document "${schemaIndexID}" with patch "${schemaPatch}"
    Then the response has status code 200 and error message ""

    And client sends request to "https://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    # Test invalid file index document (with missing basePath)
    Given variable "invalidIndexFile" is assigned the JSON value '{"fileIndex":{"basePath":""}}'
    When client sends request to "https://localhost:48426/file/operations" to create document with content "${invalidIndexFile}" in namespace "file:idx"
    Then the response has status code 500 and error message "missing base path"

  @duplicate_create_operation
  Scenario: Attempt to create the same index file on Sidetree twice. The second create operation should be rejected by the Observer.
    And the authorization bearer token for "GET" requests to path "/file" is set to "${content_r}"
    And the authorization bearer token for "POST" requests to path "/file" is set to "${content_w}"

    # Create the /content file index Sidetree document
    Given variable "contentIndexFile" is assigned the JSON value '{"fileIndex":{"basePath":"/content"}}'
    When client sends request to "https://localhost:48426/file/operations" to create document with content "${contentIndexFile}" in namespace "file:idx"
    Then the ID of the returned document is saved to variable "contentIdxID"
    Then we wait 10 seconds

    When an HTTP GET is sent to "https://localhost:48326/file/identifiers/${contentIdxID}"
    Then the JSON path "didDocument.id" of the response equals "${contentIdxID}"

    # Attempt to create the /content file index Sidetree document again
    When client sends request to "https://localhost:48426/file/operations" to create document with content "${contentIndexFile}" in namespace "file:idx"
    Then we wait 10 seconds

    # The Observer should have rejected the second create and the document resolver will not error out because of an invalid operation in the store
    When an HTTP GET is sent to "https://localhost:48326/file/identifiers/${contentIdxID}"
    Then the JSON path "didDocument.id" of the response equals "${contentIdxID}"

  @filehandler_unauthorized
  Scenario: Attempt to access the various content endpoints without providing a valid auth token
    When an HTTP GET is sent to "https://localhost:48428/internal/some-unknown-file.txt" and the returned status code is 401
    When an HTTP POST is sent to "https://localhost:48428/internal" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 401

    # Now provide valid tokens for /internal
    Given the authorization bearer token for "GET" requests to path "/internal" is set to "${content_r}"
    And the authorization bearer token for "POST" requests to path "/internal" is set to "${content_w}"
    When an HTTP GET is sent to "https://localhost:48428/internal/some-unknown-file.txt" and the returned status code is 404
    When an HTTP POST is sent to "https://localhost:48428/internal" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 400

    # Don't provide a token for POST requests to /file - should get 401
    When an HTTP GET is sent to "https://localhost:48428/file/identifiers/file:idx:1234" and the returned status code is 401

    # Now provide a valid tokens for /file
    Given the authorization bearer token for "GET" requests to path "/file" is set to "${content_r}"
    When an HTTP GET is sent to "https://localhost:48428/file/identifiers/file:idx:1234" and the returned status code is 404

  @filehandler_version
  Scenario: Version and protocol parameters
    # Protocol at time (block number) 50
    When an HTTP GET is sent to "https://localhost:48326/file/version?time=50"
    And the JSON path "version" of the response equals "0.1.1"
    And the JSON path "genesisTime" of the numeric response equals "20"
    And the JSON path "multihashAlgorithms" of the array response is not empty
    And the JSON path "maxOperationCount" of the numeric response equals "30"
    And the JSON path "maxOperationSize" of the numeric response equals "2400"
    And the JSON path "maxDeltaSize" of the numeric response equals "1700"
    And the JSON path "maxOperationHashLength" of the numeric response equals "100"
    And the JSON path "maxCasUriLength" of the numeric response equals "100"
    And the JSON path "maxMemoryDecompressionFactor" of the numeric response equals "3"
    And the JSON path "maxCoreIndexFileSize" of the numeric response equals "1000000"
    And the JSON path "maxProofFileSize" of the numeric response equals "1000000"
    And the JSON path "maxProvisionalIndexFileSize" of the numeric response equals "1000000"
    And the JSON path "maxChunkFileSize" of the numeric response equals "10000000"
    And the JSON path "compressionAlgorithm" of the response equals "GZIP"
    And the JSON path "patches" of the array response is not empty
    And the JSON path "signatureAlgorithms" of the array response is not empty
    And the JSON path "keyAlgorithms" of the array response is not empty
