#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Building fabric-cli..."

mkdir -p .build/bin
cd .build
rm -rf fabric-cli-ext

git clone https://github.com/trustbloc/fabric-cli-ext.git
cd fabric-cli-ext
git checkout $FABRIC_CLI_EXT_VERSION

make plugins

cp ./.build/bin/fabric ../bin/
cp -r ./.build/ledgerconfig/ ../ledgerconfig/
cp -r ./.build/file/ ../file/

cd ..
rm -rf ./fabric-cli-ext
