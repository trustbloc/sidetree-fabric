#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running sidetree-fabric integration tests..."
cd test/bddtests/
go test -count=1 -v -cover . -p 1 -timeout=30m
cd $PWD