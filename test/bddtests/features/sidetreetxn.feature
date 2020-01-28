#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@sidetreetxn
Feature:
  Background: Setup
    Given DCAS collection config "dcas-mychannel" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given DCAS collection config "docs-mychannel" is defined for collection "docs" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given off-ledger collection config "meta_data_coll" is defined for collection "meta_data" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=60m

    Given the channel "mychannel" is created and all peers have joined

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-mychannel"
    And "system" chaincode "document_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "docs-mychannel,meta_data_coll"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"

    And we wait 10 seconds

    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org2-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 3 seconds

  @sidetree_1
  Scenario: Sidetree Txn Test
    # sidetree content test
    When client writes content "Hello World" using "sidetreetxn_cc" on the "mychannel" channel
    Then client verifies that written content at the returned address from "sidetreetxn_cc" matches original content on the "mychannel" channel

    # document content test
    When client creates document with ID "did:sidetree:abc" using "document_cc" on the "mychannel" channel
    And we wait 5 seconds
    Then client verifies that query by index ID "did:sidetree:abc" from "document_cc" will return "1" versions of the document on the "mychannel" channel

    # Bring down peer1.org1 so that it doesn't get the documents via Gossip broadcast
    Given container "peer1.org1.example.com" is stopped
    # Wait a while so that Discovery will give up on this peer and remove it from the list of 'alive' peers
    And we wait 60 seconds

    # write sidetree transaction
    When client writes operations batch file and anchor file for ID "did:sidetree:123abc" using "sidetreetxn_cc" on the "mychannel" channel
    # Wait a while before starting peer1.org1 so that Gossip gives up trying to push the documents to the peer
    And we wait 65 seconds
    Then container "peer1.org1.example.com" is started
    # Wait a while to give peer1.org1 a chance to commit all blocks and get back in Discovery's list of 'alive' peers
    And we wait 15 seconds

    # Make sure that all peers have the document, including peer1.org1 which just came up
    Then client verifies that query by index ID "did:sidetree:123abc" from "document_cc" will return "2" versions of the document on the "mychannel" channel on peers "peer0.org1.example.com"
    And client verifies that query by index ID "did:sidetree:123abc" from "document_cc" will return "2" versions of the document on the "mychannel" channel on peers "peer0.org2.example.com"
    And client verifies that query by index ID "did:sidetree:123abc" from "document_cc" will return "2" versions of the document on the "mychannel" channel on peers "peer1.org2.example.com"
    And client verifies that query by index ID "did:sidetree:123abc" from "document_cc" will return "2" versions of the document on the "mychannel" channel on peers "peer1.org1.example.com"
