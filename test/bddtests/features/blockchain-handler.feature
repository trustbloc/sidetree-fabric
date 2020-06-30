#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@blockchain-handler
Feature:
  Background: Setup
    Given DCAS collection config "dcas-cfg" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=
    Given off-ledger collection config "diddoc-cfg" is defined for collection "diddoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "fileidx-cfg" is defined for collection "fileidxdoc" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=
    Given off-ledger collection config "meta-data-cfg" is defined for collection "meta_data" as policy="OR('IMPLICIT-ORG.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=

    Given variable "blockchain_r" is assigned the value "TOKEN_BLOCKCHAIN_R"
    And variable "cas_r" is assigned the value "TOKEN_CAS_R"
    And variable "did_w" is assigned the value "TOKEN_DID_W"

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

    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @blockchain_handler
  Scenario: Blockchain functions
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/blockchain" is set to "${blockchain_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/blockchain" is set to "${blockchain_r}"
    And the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/cas" is set to "${cas_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/operations" is set to "${did_w}"
    And the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/identifiers" is set to "${did_w}"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/version"
    Then the JSON path "name" of the response equals "Hyperledger Fabric"
    And the JSON path "version" of the response equals "2.1.1"

    Given variable "peer0.org1" is assigned the value "https://localhost:48326/sidetree/0.0.1"
    And variable "peer1.org1" is assigned the value "https://localhost:48327/sidetree/0.0.1"
    And variable "peer2.org1" is assigned the value "https://localhost:48328/sidetree/0.0.1"
    And variable "peer0.org2" is assigned the value "https://localhost:48426/sidetree/0.0.1"
    And variable "peer1.org2" is assigned the value "https://localhost:48427/sidetree/0.0.1"
    And variable "peer2.org2" is assigned the value "https://localhost:48428/sidetree/0.0.1"

    # Write several Sidetree transactions. Scatter the requests across different endpoints to generate multiple
    # Sidetree transactions within the same block and across multiple blocks.
    When client sends request to "${peer0.org1}/operations,${peer1.org1}/operations,${peer2.org1}/operations,${peer0.org2}/operations,${peer1.org2}/operations,${peer2.org2}/operations" to create 50 DID documents using 10 concurrent requests
    Then we wait 20 seconds
    Then client sends request to "${peer0.org1}/identifiers,${peer1.org1}/identifiers,${peer2.org1}/identifiers,${peer0.org2}/identifiers,${peer1.org2}/identifiers,${peer2.org2}/identifiers" to verify the DID documents that were created

    When client sends request to "${peer0.org1}/operations,${peer1.org1}/operations,${peer2.org1}/operations,${peer0.org2}/operations,${peer1.org2}/operations,${peer2.org2}/operations" to create 50 DID documents using 10 concurrent requests
    Then we wait 20 seconds
    Then client sends request to "${peer0.org1}/identifiers,${peer1.org1}/identifiers,${peer2.org1}/identifiers,${peer0.org2}/identifiers,${peer1.org2}/identifiers,${peer2.org2}/identifiers" to verify the DID documents that were created

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time"
    Then the JSON path "time" of the response is not empty
    And the JSON path "hash" of the response is not empty
    And the JSON path "previous_hash" of the response is not empty
    And the JSON path "time" of the response is saved to variable "latest-time"
    And the JSON path "hash" of the response is saved to variable "latest-hash"
    And the JSON path "previous_hash" of the response is saved to variable "latest-previous-hash"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/${latest-hash}"
    Then the JSON path "hash" of the response equals "${latest-hash}"
    And the JSON path "previous_hash" of the response equals "${latest-previous-hash}"
    And the JSON path "time" of the response equals "${latest-time}"

    # Invalid hash - Bad Request (400)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/xxx_xxx" and the returned status code is 400

    # Hash not found - Not Found (404)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/AQIDBAUGBwgJCgsM" and the returned status code is 404

    # The config setting for maxTransactionsInResponse is 10 so we should expect 10 transactions in the query for all transactions
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions"
    And the JSON path "more_transactions" of the boolean response equals "true"
    And the JSON path "transactions.0.transaction_time_hash" of the response is not empty
    And the JSON path "transactions.0.anchor_string" of the response is not empty
    And the JSON path "transactions.0.transaction_time" of the numeric response is saved to variable "time_0"
    And the JSON path "transactions.0.transaction_time_hash" of the response is saved to variable "timeHash_0"
    And the JSON path "transactions.0.transaction_number" of the numeric response is saved to variable "txnNum_0"
    And the JSON path "transactions.0.anchor_string" of the response is saved to variable "anchor_string_0"
    And the JSON path "transactions.9.transaction_time_hash" of the response is not empty
    And the JSON path "transactions.9.anchor_string" of the response is not empty
    And the JSON path "transactions.9.transaction_time" of the numeric response is saved to variable "time_9"
    And the JSON path "transactions.9.transaction_time_hash" of the response is saved to variable "timeHash_9"
    And the JSON path "transactions.9.transaction_number" of the numeric response is saved to variable "txnNum_9"
    And the JSON path "transactions.9.anchor_string" of the response is saved to variable "anchor_string_9"
    And anchor address is parsed from anchor string "anchor_string_9" and saved to variable "anchor_address_9"

    # Get more transactions from where we left off
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=${txnNum_9}&transaction-time-hash=${timeHash_9}"
    And the JSON path "transactions.0.transaction_time" of the numeric response equals "${time_9}"
    And the JSON path "transactions.0.transaction_time_hash" of the response equals "${timeHash_9}"
    And the JSON path "transactions.0.transaction_number" of the numeric response equals "${txnNum_9}"
    And the JSON path "transactions.0.anchor_string" of the response equals "${anchor_string_9}"
    And the JSON path "transactions" of the raw response is saved to variable "transactions"

    # Ensure that the anchor hash resolves to a valid value stored in DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${anchor_address_9}?max-size=1000000"
    And response is decompressed using "GZIP"
    And the JSON path "mapFileHash" of the response is saved to variable "mapFileHash"

    # Ensure that the batch file hash resolves to a valid value stored in DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${mapFileHash}?max-size=1000000"
    And response is decompressed using "GZIP"
    And the JSON path "chunks" of the array response is not empty

    # Invalid since
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=xxx&transaction-time-hash=${timeHash_9}" and the returned status code is 400
    And the JSON path "code" of the response equals "invalid_transaction_number_or_time_hash"

    # Invalid time hash
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=0&transaction-time-hash=xxx_xxx" and the returned status code is 400
    And the JSON path "code" of the response equals "invalid_transaction_number_or_time_hash"

    # Hash not found
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=0&transaction-time-hash=AQIDBAUGBwgJCgsM" and the returned status code is 404

    # Valid transactions
    When an HTTP POST is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/first-valid" with content "${transactions}" of type "application/json"
    Then the JSON path "transaction_time" of the numeric response equals "${time_9}"
    And the JSON path "transaction_time_hash" of the response equals "${timeHash_9}"
    And the JSON path "transaction_number" of the numeric response equals "${txnNum_9}"
    And the JSON path "anchor_string" of the response equals "${anchor_string_9}"

    # Invalid transactions
    Given variable "invalidTransactions" is assigned the JSON value '[{"transaction_number":3,"transaction_time":10,"transaction_time_hash":"xsZhH8Wpg5_DNEIB3KN9ihtkVuBDLWWGJ2OlVWTIZBs=","anchorString":"invalid"}]'
    When an HTTP POST is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/first-valid" with content "${invalidTransactions}" of type "application/json" and the returned status code is 404

    # Retrieve the anchor file from the transaction time in transaction 0 above
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks?from-time=${time_0}&max-blocks=2"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.header.number" of the response equals "${time_0}"
    And the JSON path "1.header.previous_hash" of the response is saved to variable "previous-hash"
    And the JSON path "1.data.data.0.payload.data.actions.0.payload.action.proposal_response_payload.extension.results.ns_rwset.1.rwset.writes.0.value" of the response is saved to variable "txn-info"
    # Binary values in the JSON block are returned as strings encoded in base64 (standard) encoding. Decoding the value will give us the (base64URL-encoded) anchor string.
    Given the base64-encoded value "${txn-info}" is decoded and saved to variable "url-encoded-txn-info"
    And anchor address is parsed from transaction info "url-encoded-txn-info" and saved to variable "url-encoded-anchor-address"
    # Retrieve the anchor file from DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${url-encoded-anchor-address}?max-size=1000000"
    And response is decompressed using "GZIP"
    Then the JSON path "mapFileHash" of the response is not empty

    # Retrieve the previous block using the previous hash from above
    # Binary values in the JSON block are returned as strings encoded in base64 (standard) encoding. Convert the string to base64URL-encoding.
    Given the base64-encoded value "${previous-hash}" is converted to base64URL-encoding and saved to variable "url-encoded-previous-hash"
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks/${url-encoded-previous-hash}"
    And the JSON path "0.header.number" of the response equals "${time_0}"

    # Retrieve the anchor file from the current block hash
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks/${latest-hash}"
    Then the JSON path "0.data.data.0.payload.data.actions.0.payload.action.proposal_response_payload.extension.results.ns_rwset.1.rwset.writes.0.value" of the response is saved to variable "txn-info"
    # Binary values in the JSON block are returned as strings encoded in base64 (standard) encoding. Decoding the value will give us the (base64URL-encoded) anchor string.
    Given the base64-encoded value "${txn-info}" is decoded and saved to variable "url-encoded-txn-info"
    And anchor address is parsed from transaction info "url-encoded-txn-info" and saved to variable "url-encoded-anchor-address"

    # Retrieve the anchor file from DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${url-encoded-anchor-address}?max-size=1000000"
    And response is decompressed using "GZIP"
    Then the JSON path "mapFileHash" of the response is not empty

    # Retrieve the anchor file from the previous block hash
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks/${latest-previous-hash}"
    Then the JSON path "0.data.data.0.payload.data.actions.0.payload.action.proposal_response_payload.extension.results.ns_rwset.1.rwset.writes.0.value" of the response is saved to variable "txn-info"
    # Binary values in the JSON block are returned as strings encoded in base64 (standard) encoding. Decoding the value will give us the (base64URL-encoded) anchor string.
    Given the base64-encoded value "${txn-info}" is decoded and saved to variable "url-encoded-txn-info"
    And anchor address is parsed from transaction info "url-encoded-txn-info" and saved to variable "url-encoded-anchor-address"
    # Retrieve the anchor file from DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${url-encoded-anchor-address}?max-size=1000000"
    And response is decompressed using "GZIP"
    Then the JSON path "mapFileHash" of the response is not empty

    # Get block by hash where the data is base64-encoded
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks/${url-encoded-previous-hash}?data-encoding=base64"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.header.number" of the numeric response equals "${time_0}"
    And the JSON path "0.header.data_hash" of the response is saved to variable "data-hash"
    And the JSON path "0.data" of the response is saved to variable "block-data"
    Then the hash of the base64-encoded value "${block-data}" equals "${data-hash}"

    # Get block by hash where the data is base64URL-encoded
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks/${url-encoded-previous-hash}?data-encoding=base64url"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.header.number" of the numeric response equals "${time_0}"
    And the JSON path "0.header.data_hash" of the response is saved to variable "data-hash"
    And the JSON path "0.data" of the response is saved to variable "block-data"
    Then the hash of the base64URL-encoded value "${block-data}" equals "${data-hash}"

    # Get blocks in range where the data is base64-encoded
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks?from-time=${time_0}&max-blocks=2&data-encoding=base64"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.header.number" of the numeric response equals "${time_0}"
    And the JSON path "0.header.data_hash" of the response is saved to variable "data-hash"
    And the JSON path "0.data" of the response is saved to variable "block-data"
    Then the hash of the base64-encoded value "${block-data}" equals "${data-hash}"

    # Get blocks in range where the data is base64URL-encoded
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/blocks?from-time=${time_0}&max-blocks=2&data-encoding=base64url"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.header.number" of the numeric response equals "${time_0}"
    And the JSON path "0.header.data_hash" of the response is saved to variable "data-hash"
    And the JSON path "0.data" of the response is saved to variable "block-data"
    Then the hash of the base64URL-encoded value "${block-data}" equals "${data-hash}"

    # Retrieve the config block for the given block number and make check the channel header type
    # to ensure the block is a config block. Following are the channel header types:
    #  MESSAGE              : 0
    #  CONFIG               : 1
    #  CONFIG_UPDATE        : 2
    #  ENDORSER_TRANSACTION : 3
    #  ORDERER_TRANSACTION  : 4
    #  DELIVER_SEEK_INFO    : 5
    #  CHAINCODE_PACKAGE    : 6
    #  PEER_ADMIN_OPERATION : 8

    # Get the latest config block (the data is in JSON format)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block"
    And the JSON path "data.data.0.payload.header.channel_header.type" of the numeric response equals "1"
    # Get the latest config block (the data is base64-encoded)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block?data-encoding=base64"
    And the JSON path "header.data_hash" of the response is saved to variable "config-data-hash"
    And the JSON path "data" of the response is saved to variable "config-data"
    Then the hash of the base64-encoded value "${config-data}" equals "${config-data-hash}"
    # Get the latest config block (the data is base64URL-encoded)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block?data-encoding=base64url"
    And the JSON path "header.data_hash" of the response is saved to variable "config-data-hash"
    And the JSON path "data" of the response is saved to variable "config-data"
    Then the hash of the base64URL-encoded value "${config-data}" equals "${config-data-hash}"

    # Get the config block that was used by the block with the given hash (the data is in JSON format)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block/${url-encoded-previous-hash}"
    And the JSON path "data.data.0.payload.header.channel_header.type" of the numeric response equals "1"
    # Get the config block that was used by the block with the given hash (the data is base64-encoded)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block/${url-encoded-previous-hash}?data-encoding=base64"
    And the JSON path "header.data_hash" of the response is saved to variable "config-data-hash"
    And the JSON path "data" of the response is saved to variable "config-data"
    Then the hash of the base64-encoded value "${config-data}" equals "${config-data-hash}"
    # Get the config block that was used by the block with the given hash (the data is base64URL-encoded)
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/config-block/${url-encoded-previous-hash}?data-encoding=base64url"
    And the JSON path "header.data_hash" of the response is saved to variable "config-data-hash"
    And the JSON path "data" of the response is saved to variable "config-data"
    Then the hash of the base64URL-encoded value "${config-data}" equals "${config-data-hash}"

  @invalid_blockchain_config
  Scenario: Invalid configuration
    Given fabric-cli context "mychannel" is used
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-blockchainhandler-config.json --noprompt" then the error response should contain "component name must be set to the base path [/sidetree/0.0.1/blockchain]"

  @blockchain_unauthorized
  Scenario: Attempt to access the blockchain endpoints without providing an auth token
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/version" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/time" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/time/hash1234" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/transactions" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/transactions?since=0&transaction-time-hash=hash1234" and the returned status code is 401
    When an HTTP POST is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/first-valid" with content "transactions" of type "application/json" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/blocks?from-time=1&max-blocks=2" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/config-block" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/config-block/hash1234" and the returned status code is 401

    # Now provide a valid token
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/blockchain" is set to "${blockchain_r}"
    Given the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/blockchain" is set to "${blockchain_r}"

    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/version" and the returned status code is 200
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/time" and the returned status code is 200
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/time/hash1234" and the returned status code is 404
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/transactions" and the returned status code is 200
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/transactions?since=0&transaction-time-hash=hash1234" and the returned status code is 404
    When an HTTP POST is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/first-valid" with content "transactions" of type "application/json" and the returned status code is 400
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/blocks?from-time=1&max-blocks=1" and the returned status code is 200
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/config-block" and the returned status code is 200
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/blockchain/config-block/hash1234" and the returned status code is 404
