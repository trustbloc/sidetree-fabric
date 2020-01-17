/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {defineSupportCode} = require('cucumber');

defineSupportCode(function ({And, But, Given, Then, When}) {
    And(/^variable "([^"]*)" is assigned config from file "([^"]*)"$/, function (arg1, arg2, callback) {
        callback.pending();
    });
});