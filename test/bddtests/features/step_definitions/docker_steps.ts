/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {defineSupportCode} = require('cucumber');

defineSupportCode(function ({And, But, Given, Then, When}) {
    this.When(/^container "([^"]*)" is started$/, function (arg1, callback) {
        callback.pending();
    });
    this.When(/^container "([^"]*)" is stopped$/, function (arg1, callback) {
        callback.pending();
    });
    this.When(/^container "([^"]*)" is paused$/, function (arg1, callback) {
        callback.pending();
    });
    this.When(/^container "([^"]*)" is unpaused$/, function (arg1, callback) {
        callback.pending();
    });
});
