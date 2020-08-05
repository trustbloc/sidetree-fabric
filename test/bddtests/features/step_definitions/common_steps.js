/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {defineSupportCode} = require('cucumber');

defineSupportCode(function ({And, But, Given, Then, When}) {
    Given(/^the channel "([^"]*)" is created and all peers have joined$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^collection config "([^"]*)" is defined for collection "([^"]*)" as policy="([^"]*)", requiredPeerCount=(\d+), maxPeerCount=(\d+), and blocksToLive=(\d+)$/, function (arg1, arg2, arg3, arg4, arg5, arg6, callback) {
        callback.pending();
    });
    Given(/^"([^"]*)" chaincode "([^"]*)" is installed from path "([^"]*)" to all peers$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    Given(/^"([^"]*)" chaincode "([^"]*)" is instantiated from path "([^"]*)" on the "([^"]*)" channel with args "([^"]*)" with endorsement policy "([^"]*)" with collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, callback) {
        callback.pending();
    });
    Given(/^chaincode "([^"]*)" is warmed up on all peers on the "([^"]*)" channel$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    When(/^client queries chaincode "([^"]*)" with args "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    Then(/^response from "([^"]*)" to client equal value "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Given(/^we wait (\d+) seconds$/, function (arg1, callback) {
        callback.pending();
    });
    When(/^the response is saved to variable "([^"]*)"$/, function (arg1, callback) {
        callback.pending();
    });
    When(/^client queries chaincode "([^"]*)" with args "([^"]*)" on the "([^"]*)" channel then the error response should contain "([^"]*)"$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    When(/^client queries chaincode "([^"]*)" with args "([^"]*)" on a single peer in the "([^"]*)" org on the "([^"]*)" channel$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    When(/^client queries chaincode "([^"]*)" with args "([^"]*)" on peers "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    When(/^client invokes chaincode "([^"]*)" with args "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    When(/^client invokes chaincode "([^"]*)" with args "([^"]*)" on peers "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    Given(/^"([^"]*)" chaincode "([^"]*)" version "([^"]*)" is installed from path "([^"]*)" to all peers$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    Given(/^"([^"]*)" chaincode "([^"]*)" is upgraded with version "([^"]*)" from path "([^"]*)" on the "([^"]*)" channel with args "([^"]*)" with endorsement policy "([^"]*)" with collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, callback) {
        callback.pending();
    });
    Given(/^"([^"]*)" chaincode "([^"]*)" is upgraded with version "([^"]*)" from path "([^"]*)" on the "([^"]*)" channel with args "([^"]*)" with endorsement policy "([^"]*)" with collection policy "([^"]*)" then the error response should contain "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, callback) {
        callback.pending();
    });
    Given(/^variable "([^"]*)" is assigned the JSON value '([^']*)'$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response has (\d+) items$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response equals "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the numeric response equals "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the boolean response equals "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response contains "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response does not contain "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response is saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the numeric response is saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the boolean response is saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the raw response is saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the response is not empty$/, function (arg1, callback) {
        callback.pending();
    });
    Then(/^the JSON path "([^"]*)" of the array response is not empty$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^an HTTP GET is sent to "([^"]*)"$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^an HTTP GET is sent to "([^"]*)" and the returned status code is (\d+)$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^an HTTP POST is sent to "([^"]*)" with content from file "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^an HTTP POST is sent to "([^"]*)" with content from file "([^"]*)" and the returned status code is (\d+)$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    And(/^an HTTP POST is sent to "([^"]*)" with content "([^"]*)" of type "([^"]*)"$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    And(/^an HTTP POST is sent to "([^"]*)" with content "([^"]*)" of type "([^"]*)" and the returned status code is (\d+)$/, function (arg1, arg2, arg3, arg4, callback) {
        callback.pending();
    });
    Then(/^the response equals "([^"]*)"$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^the base64-encoded value "([^"]*)" is converted to base64URL-encoding and saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^the base64-encoded value "([^"]*)" is decoded and saved to variable "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^the value "([^"]*)" equals "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^variable "([^"]*)" is assigned the value "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    And(/^the authorization bearer token for "([^"]*)" requests to path "([^"]*)" is set to "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });

    // Lifecycle steps
    Given(/^chaincode "([^"]*)" is installed from path "([^"]*)" to all peers$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    When(/^chaincode "([^"]*)", version "([^"]*)", package ID "([^"]*)", sequence (\d+) is approved by orgs "([^"]*)" on the "([^"]*)" channel with endorsement policy "([^"]*)" and collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, callback) {
        callback.pending();
    });
    When(/^chaincode "([^"]*)", version "([^"]*)", package ID "([^"]*)", sequence (\d+) is approved by orgs "([^"]*)" on the "([^"]*)" channel with endorsement policy "([^"]*)" and collection policy "([^"]*)" then the error response should contain "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, callback) {
        callback.pending();
    });
    When(/^chaincode "([^"]*)", version "([^"]*)", sequence (\d+) is committed by orgs "([^"]*)" on the "([^"]*)" channel with endorsement policy "([^"]*)" and collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, callback) {
        callback.pending();
    });
    When(/^chaincode "([^"]*)", version "([^"]*)", package ID "([^"]*)", sequence (\d+) is checked for readiness by orgs "([^"]*)" on the "([^"]*)" channel with endorsement policy "([^"]*)" and collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, callback) {
        callback.pending();
    });
    When(/^peer "([^"]*)" is queried for installed chaincodes$/, function (arg1, callback) {
        callback.pending();
    });
    When(/^committed chaincode "([^"]*)" is queried by orgs "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    When(/^all committed chaincodes are queried by orgs "([^"]*)" on the "([^"]*)" channel$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    When(/^peer "([^"]*)" is queried for approved chaincode "([^"]*)" and sequence (\d+) on the "([^"]*)" channel$/, function (arg1, arg2, arg4, arg4, callback) {
        callback.pending();
    });
    When(/^peer "([^"]*)" is queried for installed chaincode package "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
    When(/^chaincode "([^"]*)", version "([^"]*)", package ID "([^"]*)", sequence (\d+) is approved and committed by orgs "([^"]*)" on the "([^"]*)" channel with endorsement policy "([^"]*)" and collection policy "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, callback) {
        callback.pending();
    });
});
