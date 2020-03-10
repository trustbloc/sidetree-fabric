/*
    Copyright SecureKey Technologies Inc. All Rights Reserved.

    SPDX-License-Identifier: Apache-2.0
*/

var {When} = require('cucumber');
var myStepDefinitionsWrapper = function () {
    this.When(/^client sends request to "([^"]*)" to create document "([^"]*)" in namespace "([^"]*)"$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    this.When(/^client sends request to "([^"]*)" to create document with content "([^"]*)" in namespace "([^"]*)"$/, function (arg1, arg2, arg3, callback) {
        callback.pending();
    });
    this.When(/^client sends request to "([^"]*)" to retrieve file$/, function (callback) {
        callback.pending();
    });
    this.When(/^client sends request to "([^"]*)" to upload file "([^"]*)" with content type "([^"]*)"$/, function (callback) {
        callback.pending();
    });
    this.When(/^the ID of the file is saved to variable "([^"]*)"/, function (callback) {
        callback.pending();
    });
    this.When(/^the ID of the returned document is saved to variable "([^"]*)"/, function (callback) {
        callback.pending();
    });
    this.When(/^the retrieved file contains "([^"]*)"/, function (callback) {
        callback.pending();
    });
    this.When(/^the response has status code (\d+) and error message "([^"]*)"/, function (callback) {
        callback.pending();
    });
};
module.exports = myStepDefinitionsWrapper;
