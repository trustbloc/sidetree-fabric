#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

declare -a arr=(
   "cmd/chaincode/txn"
   "cmd/chaincode/doc"
)

function finish {
  rm -rf vendor
}
trap finish EXIT

go mod vendor


echo "Copy cc..."
for i in "${arr[@]}"
do

mkdir -p ./.build/cc/src/github.com/trustbloc/sidetree-fabric/"${i//./}"
cp -r $i/* ./.build/cc/src/github.com/trustbloc/sidetree-fabric/"${i//./}"
find ./vendor ! -name '*_test.go' | cpio -pdm ./.build/cc/src/github.com/trustbloc/sidetree-fabric/"${i//./}"
mkdir -p ./.build/cc/src/github.com/trustbloc/sidetree-fabric/"${i//./}"/vendor/github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas
cp -r ./cmd/chaincode/cas/* ./.build/cc/src/github.com/trustbloc/sidetree-fabric/"${i//./}"/vendor/github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas

done
