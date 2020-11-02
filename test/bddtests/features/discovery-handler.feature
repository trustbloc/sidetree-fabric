#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@discovery-handler
Feature:
  Background: Setup
    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""

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

  @discovery_handler
  Scenario: Discovery queries
    Given the authorization bearer token for "GET" requests to path "/discovery/0.0.1" is set to "TOKEN_DISCOVERY_R"

    When an HTTP GET is sent to "https://localhost:48326/discovery/0.0.1"
    And the JSON path "#.service" of the response contains "did:sidetree"
    And the JSON path "#.service" of the response contains "cas"
    And the JSON path "#.service" of the response contains "blockchain"
    And the JSON path "#.domain" of the response contains "org1.example.com"
    And the JSON path "#.domain" of the response contains "org2.example.com"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1/blockchain"

    # Filter by service
    When an HTTP GET is sent to "https://localhost:48326/discovery/0.0.1?service=did:sidetree&service=cas"
    And the JSON path "#.service" of the response contains "did:sidetree"
    And the JSON path "#.service" of the response contains "cas"
    And the JSON path "#.service" of the response does not contain "blockchain"
    And the JSON path "#.domain" of the response contains "org2.example.com"
    And the JSON path "#.domain" of the response contains "org1.example.com"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org2.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org2.example.com:48326/sidetree/0.0.1/blockchain"

    # Filter by service and domain
    When an HTTP GET is sent to "https://localhost:48326/discovery/0.0.1?service=did:sidetree&service=cas&domain=org2.example.com"
    And the JSON path "#.service" of the response contains "did:sidetree"
    And the JSON path "#.service" of the response contains "cas"
    And the JSON path "#.service" of the response does not contain "blockchain"
    And the JSON path "#.domain" of the response contains "org2.example.com"
    And the JSON path "#.domain" of the response does not contain "org1.example.com"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org1.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org1.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer0.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response contains "https://peer1.org2.example.com:48326/sidetree/0.0.1/cas"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org1.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer0.org2.example.com:48326/sidetree/0.0.1/blockchain"
    And the JSON path "#.rootEndpoint" of the response does not contain "https://peer1.org2.example.com:48326/sidetree/0.0.1/blockchain"
