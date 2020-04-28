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
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/version"
    Then the JSON path "name" of the response equals "Hyperledger Fabric"
    And the JSON path "version" of the response equals "2.0.0"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time"
    Then the JSON path "time" of the response is not empty
    And the JSON path "hash" of the response is not empty
    And the JSON path "time" of the response is saved to variable "time"
    And the JSON path "hash" of the response is saved to variable "hash"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/${hash}"
    Then the JSON path "hash" of the response equals "${hash}"
    And the JSON path "time" of the response equals "${time}"

    # Invalid hash - Bad Request (400)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/xxx_xxx" and the returned status code is 400

    # Hash not found - Not Found (404)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/time/AQIDBAUGBwgJCgsM" and the returned status code is 404

    # Write a few Sidetree transactions. Scatter the requests across different endpoints to generate multiple
    # Sidetree transactions within the same block. The Orderer's batch timeout is set to 2s, so sleep 2s between
    # writes to guarantee that we generate a few blocks.
    Then client sends request to "https://localhost:48326/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And check success response contains "#didDocumentHash"
    And client sends request to "https://localhost:48327/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48328/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"

    Then we wait 2 seconds

    Then client sends request to "https://localhost:48427/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48428/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48326/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48327/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48328/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"

    Then we wait 2 seconds

    Then client sends request to "https://localhost:48426/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48427/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"
    And client sends request to "https://localhost:48428/sidetree/0.0.1/operations" to create DID document in namespace "did:sidetree"

    Then we wait 15 seconds

    # The config setting for maxTransactionsInResponse is 10 so we should expect 10 transactions in the query for all transactions
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions"
    And the JSON path "more_transactions" of the boolean response equals "true"
    And the JSON path "transactions.9.transaction_time_hash" of the response is not empty
    And the JSON path "transactions.9.anchor_string" of the response is not empty
    And the JSON path "transactions.9.transaction_time" of the numeric response is saved to variable "time_9"
    And the JSON path "transactions.9.transaction_time_hash" of the response is saved to variable "timeHash_9"
    And the JSON path "transactions.9.transaction_number" of the numeric response is saved to variable "txnNum_9"
    And the JSON path "transactions.9.anchor_string" of the response is saved to variable "anchor_9"

    # Get more transactions from where we left off
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=${txnNum_9}&transaction_time_hash=${timeHash_9}"
    And the JSON path "transactions.0.transaction_time" of the numeric response equals "${time_9}"
    And the JSON path "transactions.0.transaction_time_hash" of the response equals "${timeHash_9}"
    And the JSON path "transactions.0.transaction_number" of the numeric response equals "${txnNum_9}"
    And the JSON path "transactions.0.anchor_string" of the response equals "${anchor_9}"
    And the JSON path "transactions" of the raw response is saved to variable "transactions"

    # Ensure that the anchor hash resolves to a valid value stored in DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${anchor_9}?max_size=1024"
    And the JSON path "batchFileHash" of the response is saved to variable "batchFileHash"

    # Ensure that the batch file hash resolves to a valid value stored in DCAS
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${batchFileHash}?max_size=8192"
    And the JSON path "operations" of the array response is not empty

    # Invalid since
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=xxx&transaction_time_hash=${timeHash_9}" and the returned status code is 400
    And the JSON path "code" of the response equals "invalid_transaction_number_or_time_hash"

    # Invalid time hash
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=0&transaction_time_hash=xxx_xxx" and the returned status code is 400
    And the JSON path "code" of the response equals "invalid_transaction_number_or_time_hash"

    # Hash not found
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/transactions?since=0&transaction_time_hash=AQIDBAUGBwgJCgsM" and the returned status code is 404

    # Valid transactions
    When an HTTP POST is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/firstValid" with content "${transactions}" of type "application/json"
    Then the JSON path "transaction_time" of the numeric response equals "${time_9}"
    And the JSON path "transaction_time_hash" of the response equals "${timeHash_9}"
    And the JSON path "transaction_number" of the numeric response equals "${txnNum_9}"
    And the JSON path "anchor_string" of the response equals "${anchor_9}"

    # Invalid transactions
    Given variable "invalidTransactions" is assigned the JSON value '[{"transaction_number":3,"transaction_time":10,"transaction_time_hash":"xsZhH8Wpg5_DNEIB3KN9ihtkVuBDLWWGJ2OlVWTIZBs=","anchorString":"invalid"}]'
    When an HTTP POST is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/firstValid" with content "${invalidTransactions}" of type "application/json" and the returned status code is 404

    # Blockchain info
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/blockchain/info"
    And the JSON path "current_time_hash" of the response is not empty
    And the JSON path "previous_time_hash" of the response is not empty

  @invalid_blockchain_config
  Scenario: Invalid configuration
    Given fabric-cli context "mychannel" is used
    When fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/invalid-blockchainhandler-config.json --noprompt" then the error response should contain "component name must be set to the base path [/sidetree/0.0.1/blockchain]"
