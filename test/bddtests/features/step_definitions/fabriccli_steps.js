/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {defineSupportCode} = require('cucumber');

defineSupportCode(function ({And, But, Given, Then, When}) {
    And(/^fabric-cli network is initialized$/, function (callback) {
        callback.pending();
    });
    And(/^fabric-cli plugin "([^"]*)" is installed$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^fabric-cli context "([^"]*)" is defined on channel "([^"]*)" with org "([^"]*)", peers "([^"]*)" and user "([^"]*)"$/, function (arg1, arg2, arg3, arg4, arg5, callback) {
        callback.pending();
    });
    And(/^fabric-cli context "([^"]*)" is used$/, function (arg1, callback) {
        callback.pending();
    });
    And(/^fabric-cli is executed with args "([^"]*)"$/, function (arg1, callback) {
        callback.pending();
    });
    When(/^fabric-cli is executed with args "([^"]*)" then the error response should contain "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
});
