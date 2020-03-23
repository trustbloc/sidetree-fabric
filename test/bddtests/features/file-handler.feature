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
    And fabric-cli plugin "../../.build/file" is installed
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
    # Create the /schema file index Sidetree document
    When fabric-cli is executed with args "file createidx --path /schema --url http://localhost:48526/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt"
    And the JSON path "id" of the response is saved to variable "schemaIndexID"
    # Update the file handler configuration for the '/schema' path with the ID of the '/schema' file index document
    Then fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org1MSP --peers peer0.org1.example.com;peer1.org1.example.com --path /schema --idxid ${schemaIndexID} --noprompt"
    And fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org2MSP --peers peer0.org2.example.com;peer1.org2.example.com --path /schema --idxid ${schemaIndexID} --noprompt"

    # Create the /.well-known/did-bloc file index Sidetree document
    When fabric-cli is executed with args "file createidx --path /.well-known/did-bloc --url http://localhost:48526/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt"
    And the JSON path "id" of the response is saved to variable "wellKnownIndexID"
    # Update the file handler configuration for the '/.well-known/did-bloc' path with the ID of the '/.well-known/did-bloc' file index document
    Then fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org1MSP --peers peer0.org1.example.com;peer1.org1.example.com --path /.well-known/did-bloc --idxid ${wellKnownIndexID} --noprompt"
    And fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org2MSP --peers peer0.org2.example.com;peer1.org2.example.com --path /.well-known/did-bloc --idxid ${wellKnownIndexID} --noprompt"

    And we wait 10 seconds

    # Upload arrays schema file
    When fabric-cli is executed with args "file upload --url http://localhost:48326/schema --files ./fixtures/testdata/schemas/arrays.schema.json --idxurl http://localhost:48326/file/${schemaIndexID} --pwd pwd1 --nextpwd pwd2 --noprompt"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.Name" of the response equals "arrays.schema.json"
    And the JSON path "0.ContentType" of the response equals "application/json"

    # Upload .well-known files
    When fabric-cli is executed with args "file upload --url http://localhost:48326/.well-known/did-bloc --files ./fixtures/testdata/well-known/trustbloc.dev.json;./fixtures/testdata/well-known/org1.dev.json;fixtures/testdata/well-known/org2.dev.json --idxurl http://localhost:48326/file/${wellKnownIndexID} --pwd pwd1 --nextpwd pwd2 --noprompt"
    Then the JSON path "#" of the response has 3 items
    And the JSON path "0.Name" of the response equals "trustbloc.dev.json"
    And the JSON path "0.ContentType" of the response equals "application/json"
    And the JSON path "1.Name" of the response equals "org1.dev.json"
    And the JSON path "1.ContentType" of the response equals "application/json"
    And the JSON path "2.Name" of the response equals "org2.dev.json"
    And the JSON path "2.ContentType" of the response equals "application/json"

    # Resolve schema files
    When client sends request to "http://localhost:48326/schema/arrays.schema.json" to retrieve file
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"
    # geographical-location.schema.json should not be there until we upload it and update the index
    When client sends request to "http://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the response has status code 404 and error message "file not found"

    # Resolve .well-known files
    When client sends request to "http://localhost:48626/.well-known/did-bloc/trustbloc.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "trustbloc.dev"
    When client sends request to "http://localhost:48626/.well-known/did-bloc/org1.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "org1.dev"
    When client sends request to "http://localhost:48626/.well-known/did-bloc/org2.dev.json" to retrieve file
    Then the JSON path "domain" of the response equals "org2.dev"

    # Upload a new schema and update the schema index document
    When fabric-cli is executed with args "file upload --url http://localhost:48326/schema --files ./fixtures/testdata/schemas/geographical-location.schema.json --idxurl http://localhost:48326/file/${schemaIndexID} --pwd pwd2 --nextpwd pwd3 --noprompt"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.Name" of the response equals "geographical-location.schema.json"
    And the JSON path "0.ContentType" of the response equals "application/json"

    And client sends request to "http://localhost:48326/schema/geographical-location.schema.json" to retrieve file
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

  @duplicate_create_operation
  Scenario: Attempt to create the same index file on Sidetree twice. The second create operation should be rejected by the Observer.
    # Create the /content file index Sidetree document
    When fabric-cli is executed with args "file createidx --path /content --url http://localhost:48526/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt"
    And the JSON path "id" of the response is saved to variable "fileIdxID"
    Then we wait 10 seconds

    When an HTTP request is sent to "http://localhost:48326/file/${fileIdxID}"
    Then the JSON path "id" of the response equals "${fileIdxID}"

    # Attempt to create the /content file index Sidetree document again
    When fabric-cli is executed with args "file createidx --path /content --url http://localhost:48526/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt"
    And the JSON path "id" of the response is saved to variable "fileIdxID"
    Then we wait 10 seconds

    # The Observer should have rejected the second create and the document resolver will not error out because of an invalid operation in the store
    When an HTTP request is sent to "http://localhost:48326/file/${fileIdxID}"
    Then the JSON path "id" of the response equals "${fileIdxID}"
