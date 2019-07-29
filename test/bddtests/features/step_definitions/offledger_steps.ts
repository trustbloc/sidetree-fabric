/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {defineSupportCode} = require('cucumber');

defineSupportCode(function ({And, But, Given, Then, When}) {
    And(/^off-ledger collection config "([^"]*)" is defined for collection "([^"]*)" as policy="([^"]*)", requiredPeerCount=(\d+), maxPeerCount=(\d+), and timeToLive=([^"]*)$/, function (arg1, arg2, arg3, arg4, arg5, arg6, callback) {
        callback.pending();
    });
    And(/^DCAS collection config "([^"]*)" is defined for collection "([^"]*)" as policy="([^"]*)", requiredPeerCount=(\d+), maxPeerCount=(\d+), and timeToLive=([^"]*)$/, function (arg1, arg2, arg3, arg4, arg5, arg6, callback) {
        callback.pending();
    });
    Given(/^variable "([^"]*)" is assigned the CAS key of value "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
});
