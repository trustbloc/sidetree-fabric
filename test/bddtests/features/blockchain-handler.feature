#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@blockchain-handler
Feature:
  Background: Setup
    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"

    And we wait 10 seconds

    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-blockchainhandler-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-blockchainhandler-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @blockchain_s1
  Scenario: Blockchain functions
    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.1.3/blockchain/version"
    Then the JSON path "name" of the response equals "Hyperledger Fabric"
    And the JSON path "version" of the response equals "2.0.0"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.1.3/blockchain/time"
    Then the JSON path "time" of the response is not empty
    And the JSON path "hash" of the response is not empty
    And the JSON path "time" of the response is saved to variable "time"
    And the JSON path "hash" of the response is saved to variable "hash"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.1.3/blockchain/time/${hash}"
    And the JSON path "hash" of the response equals "${hash}"
    And the JSON path "time" of the response equals "${time}"

    # Invalid hash - Bad Request (400)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.1.3/blockchain/time/xxx_xxx" and the returned status code is 400

    # Hash not found - Not Found (404)
    Then an HTTP GET is sent to "https://localhost:48326/sidetree/0.1.3/blockchain/time/AQIDBAUGBwgJCgsM" and the returned status code is 404
