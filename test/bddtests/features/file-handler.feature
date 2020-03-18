#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@file-handler
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

    Given DCAS collection config "consortium-files-coll" is defined for collection "consortium" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    And "system" chaincode "files" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "consortium-files-coll"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"

    And we wait 10 seconds

    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @upload_and_retrieve_files
  Scenario: upload files to DCAS and create index file on Sidetree that references them
    # Upload schema files
    When client sends request to "http://localhost:48326/schema" to upload file "fixtures/testdata/schemas/arrays.schema.json" with content type "application/json"
    Then the ID of the file is saved to variable "arraysSchemaID"
    # Create the schema file index Sidetree document
    Given variable "schemaIndexFile" is assigned the JSON value '{"arrays.schema.json":"${arraysSchemaID}"}'
    When client sends request to "http://localhost:48526/file" to create document with content "${schemaIndexFile}" in namespace "file:idx"
    Then the ID of the returned document is saved to variable "schemaIndexID"

    # Upload .well-known files
    When client sends request to "http://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/trustbloc.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownTrustblocID"
    When client sends request to "http://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/org1.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownOrg1ID"
    When client sends request to "http://localhost:48326/.well-known/did-bloc" to upload file "fixtures/testdata/well-known/org2.dev.json" with content type "application/json"
    Then the ID of the file is saved to variable "wellKnownOrg2ID"
    # Create the .well-known file index Sidetree document
    Given variable "wellKnownIndexFile" is assigned the JSON value '{"trustbloc.dev.json":"${wellKnownTrustblocID}","org1.dev.json":"${wellKnownOrg1ID}","org2.dev.json":"${wellKnownOrg2ID}"}'
    When client sends request to "http://localhost:48526/file" to create document with content "${wellKnownIndexFile}" in namespace "file:idx"
    Then the ID of the returned document is saved to variable "wellKnownIndexID"

    # Update the ledger config to point to the index file documents
    Given variable "schemaHandlerConfig" is assigned the JSON value '{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"${schemaIndexID}"}'
    And variable "wellKnownHandlerConfig" is assigned the JSON value '{"BasePath":"/.well-known/did-bloc","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"${wellKnownIndexID}"}'
    And variable "org1ConfigUpdate" is assigned the JSON value '{"MspID":"Org1MSP","Peers":[{"PeerID":"peer0.org1.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer1.org1.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]}]}'
    And variable "org2ConfigUpdate" is assigned the JSON value '{"MspID":"Org2MSP","Peers":[{"PeerID":"peer0.org2.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]},{"PeerID":"peer1.org2.example.com","Apps":[{"AppName":"file-handler","Version":"1","Components":[{"Name":"/schema","Version":"1","Config":"${schemaHandlerConfig}","Format":"json"},{"Name":"/.well-known/did-bloc","Version":"1","Config":"${wellKnownHandlerConfig}","Format":"json"}]}]}]}'
    And fabric-cli is executed with args "ledgerconfig update --config ${org1ConfigUpdate} --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --config ${org2ConfigUpdate} --noprompt"

    # Resolve schema files
    When client sends request to "http://localhost:48326/schema/arrays.schema.json" to retrieve file
    Then the retrieved file contains "https://example.com/arrays.schema.json"
    # geographical-location.schema.json should not be there until we upload it and update the index
    When client sends request to "http://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the response has status code 404 and error message "file not found"

    # Resolve .well-known files
    When client sends request to "http://localhost:48626/.well-known/did-bloc/trustbloc.dev.json" to retrieve file
    Then the retrieved file contains "trustbloc.dev"
    When client sends request to "http://localhost:48626/.well-known/did-bloc/org1.dev.json" to retrieve file
    Then the retrieved file contains "org1.dev"
    When client sends request to "http://localhost:48626/.well-known/did-bloc/org2.dev.json" to retrieve file
    Then the retrieved file contains "org2.dev"

    # Upload a new schema and update the schema index document
    When client sends request to "http://localhost:48326/schema" to upload file "fixtures/testdata/schemas/geographical-location.schema.json" with content type "application/json"
    Then the ID of the file is saved to variable "locationsSchemaID"
    # Update the schema file index Sidetree document
    Given variable "schemaPatch" is assigned the JSON patch '[{"op": "add", "path": "/geographical-location.schema.json", "value": "${locationsSchemaID}"}]'
    When client sends request to "http://localhost:48326/file" to update document "${schemaIndexID}" with patch "${schemaPatch}"

    And client sends request to "http://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the retrieved file contains "https://example.com/geographical-location.schema.json"
