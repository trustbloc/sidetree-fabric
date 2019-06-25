#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@sidetreetxn
Feature:

    @sidetree_1
    Scenario: Sidetree Txn Test

        Given DCAS collection config "dcas-mychannel" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=6000s
        Given DCAS collection config "docs-mychannel" is defined for collection "docs" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=3, maxPeerCount=3, and timeToLive=6000s

        Given the channel "mychannel" is created and all peers have joined

        And "system" chaincode "sidetreetxn_cc" is installed from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" to all peers
        And "system" chaincode "sidetreetxn_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn" on the "mychannel" channel with args "" with endorsement policy "" with collection policy "dcas-mychannel"
        And chaincode "sidetreetxn_cc" is warmed up on all peers on the "mychannel" channel

        And "system" chaincode "document_cc" is installed from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc" to all peers
        And "system" chaincode "document_cc" is instantiated from path "github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc" on the "mychannel" channel with args "" with endorsement policy "" with collection policy "docs-mychannel"
        And chaincode "document_cc" is warmed up on all peers on the "mychannel" channel

        # sidetree content test
        When client writes content "Hello World" using "sidetreetxn_cc" on all peers in the "peerorg1" org on the "mychannel" channel
        Then client verifies that written content at the returned address from "sidetreetxn_cc" matches original content on all peers in the "peerorg1" org on the "mychannel" channel

        # document content test
        When client creates document with ID "did:sidetree:abc" using "document_cc" on all peers in the "peerorg1" org on the "mychannel" channel
        Then client verifies that query by index ID "did:sidetree:abc" from "document_cc" will return "1" versions of the document on one peer in the "peerorg1" org on the "mychannel" channel

        # write sidetree transaction
        When client writes operations batch file and anchor file for ID "did:sidetree:123abc" using "sidetreetxn_cc" on all peers in the "peerorg1" org on the "mychannel" channel
        Then client verifies that query by index ID "did:sidetree:123abc" from "document_cc" will return "2" versions of the document on one peer in the "peerorg1" org on the "mychannel" channel

