#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

declare -a testArr=(
   "e2e_cc"
)

function finish {
  rm -rf vendor
}
trap finish EXIT

go mod vendor

echo "Copy test cc..."
for i in "${testArr[@]}"
do
  mkdir -p ./.build/cc/src/github.com/trustbloc/sidetree-fabric/test/chaincode/"${i//./}"
  cp -r "test/bddtests/fixtures/testdata/chaincode/"$i/* ./.build/cc/src/github.com/trustbloc/sidetree-fabric/test/chaincode/"${i//./}"
  find ./vendor ! -name '*_test.go' | cpio -pdm ./.build/cc/src/github.com/trustbloc/sidetree-fabric/test/chaincode/"${i//./}"
done
