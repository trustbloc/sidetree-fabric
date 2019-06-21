#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

# Packages to exclude
PKGS=`go list github.com/trustbloc/sidetree-fabric/... 2> /dev/null | \
                                                   grep -v /mocks`
echo "Running pkg unit tests..."
go test -tags testing -count=1 -cover $PKGS -p 1 -timeout=10m -race -coverprofile=coverage.txt -covermode=atomic

