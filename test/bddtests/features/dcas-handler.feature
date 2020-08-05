#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@dcas-handler
Feature:
  Background: Setup
    Given DCAS collection config "dcas-cfg" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=

    Given variable "cas_r" is assigned the value "TOKEN_CAS_R"
    And variable "cas_w" is assigned the value "TOKEN_CAS_W"

    Given the channel "mychannel" is created and all peers have joined

    # Give the peers some time to gossip their new channel membership
    And we wait 20 seconds

    Then chaincode "configscc", version "v1", package ID "configscc:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy ""
    And chaincode "sidetreetxn", version "v1", package ID "sidetreetxn:v1", sequence 1 is approved and committed by orgs "peerorg1,peerorg2" on the "mychannel" channel with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" and collection policy "dcas-cfg"

    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "mychannel" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"

    And we wait 10 seconds

    Then fabric-cli context "mychannel" is used
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org1-dcashandler-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/mychannel-org2-dcashandler-config.json --noprompt"

    # Wait for the Sidetree services to start up on mychannel
    And we wait 10 seconds

  @dcas_upload_and_retrieve_content
  Scenario: Upload files to DCAS
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/cas" is set to "${cas_r}"
    And the authorization bearer token for "POST" requests to path "/sidetree/0.0.1/cas" is set to "${cas_w}"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/version"
    Then the JSON path "name" of the response equals "cas"
    And the JSON path "version" of the response equals "0.1.3"

    When an HTTP POST is sent to "https://localhost:48326/sidetree/0.0.1/cas" with content from file "fixtures/testdata/schemas/geographical-location.schema.json"
    Then the JSON path "hash" of the response is saved to variable "contentHash"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${contentHash}?max-size=1" and the returned status code is 400
    Then the response equals "content_exceeds_maximum_allowed_size"

    When an HTTP GET is sent to "https://localhost:48326/sidetree/0.0.1/cas/${contentHash}?max-size=1024"
    And the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

  @dcas_unauthorized
  Scenario: Attempt to access the cas endpoints without providing an auth token
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/cas/version" and the returned status code is 401
    When an HTTP POST is sent to "https://localhost:48428/sidetree/0.0.1/cas" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/cas/hash1234?max-size=1024" and the returned status code is 401

    # Now provide valid tokens
    Given the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/cas" is set to "${cas_r}"
    And the authorization bearer token for "GET" requests to path "/sidetree/0.0.1/cas" is set to "${cas_w}"
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/cas/version" and the returned status code is 200
    When an HTTP POST is sent to "https://localhost:48428/sidetree/0.0.1/cas" with content from file "fixtures/testdata/schemas/geographical-location.schema.json" and the returned status code is 401
    When an HTTP GET is sent to "https://localhost:48428/sidetree/0.0.1/cas/hash1234?max-size=1024" and the returned status code is 404
