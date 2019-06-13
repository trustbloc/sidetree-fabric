#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script populates the fixtures folder.

set -e

LASTRUN_CHANNEL_INFO_FILENAME="populate-channel-fixtures.txt"
LASTRUN_CRYPTO_INFO_FILENAME="populate-crypto-fixtures.txt"
FIXTURES_CHANNEL_TREE_FILENAME="fixtures-channel-tree.txt"
FIXTURES_CRYPTO_TREE_FILENAME="fixtures-crypto-tree.txt"
SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" scripts/populate-fixtures.sh)
DATE=$(date +"%m-%d-%Y")

CACHE_PATH=""
function setCachePath {
    declare envOS=$(uname -s)
    declare pkgDir="sidetree-fabric"

    if [ ${envOS} = 'Darwin' ]; then
        CACHE_PATH="${HOME}/Library/Caches/${pkgDir}"
    else
        CACHE_PATH="${HOME}/.cache/${pkgDir}"
    fi
}

# recordCacheResult writes the date and revision of successful script runs, to preempt unnecessary installs.
function recordChannelCacheResult {
    declare FIXTURES_TREE_CHANNEL=$(ls -R test/bddtests/fixtures/fabric/channel)

    mkdir -p ${CACHE_PATH}
    echo ${SCRIPT_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_CHANNEL_INFO_FILENAME}"
    echo "${FIXTURES_TREE_CHANNEL}" > "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}"
}

function recordCryptoCacheResult {
    declare FIXTURES_TREE_CRYPTO=$(ls -R test/bddtests/fixtures/fabric/crypto-config)

    mkdir -p ${CACHE_PATH}
    echo ${SCRIPT_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_CRYPTO_INFO_FILENAME}"
    echo "${FIXTURES_TREE_CRYPTO}" > "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}"
}

function isCryptoFixturesCurrent {
    if [ ! -d "test/bddtests/fixtures/fabric/crypto-config" ]; then
        echo "Crypto config directory does not exist - will need to populate fixture"
        return 1
    fi

    if [ ! -f "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}" ]; then
        echo "Fixtures crypto cache doesn't exist - populating fixtures"
        return 1
    fi

    declare FIXTURES_TREE=$(ls -R test/bddtests/fixtures/fabric/crypto-config)
    declare cachedFixturesTree=$(< "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}")
    if [ "${FIXTURES_TREE}" != "${cachedFixturesTree}" ]; then
        echo "Fixtures crypto directory modified - will need to repopulate fixtures"
        return 1
    fi
}

function isChannelFixturesCurrent {
    if [ ! -d "test/bddtests/fixtures/fabric/channel" ]; then
        echo "Channel directory does not exist - will need to populate fixture"
        return 1
    fi

    if [ ! -f "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}" ]; then
        echo "Fixtures channel cache doesn't exist - populating fixtures"
        return 1
    fi

    declare FIXTURES_TREE=$(ls -R test/bddtests/fixtures/fabric/channel)
    declare cachedFixturesTree=$(< "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}")
    if [ "${FIXTURES_TREE}" != "${cachedFixturesTree}" ]; then
        echo "Fixtures channel directory modified - will need to repopulate fixtures"
        return 1
    fi
}

function isPopulateCryptoCurrent {
    if ! isCryptoFixturesCurrent; then
        return 1
    fi
}

function isPopulateChannelCurrent {
    if ! isChannelFixturesCurrent; then
        return 1
    fi
}

function isForceMode {
    if [ "${BASH_ARGV[0]}" != "-f" ]; then
        return 1
    fi
}

function generateCryptoConfig {
    rm -Rf test/bddtests/fixtures/fabric/*/channel
    make crypto-gen
}

function generateChannelConfig {
    echo "Generating channel config ..."
    make channel-config-gen
}

setCachePath

if ! isPopulateCryptoCurrent || isForceMode; then
    generateCryptoConfig
    recordCryptoCacheResult
else
    echo "No need to populate crypto fixtures"
fi

if ! isPopulateChannelCurrent || isForceMode; then
    generateChannelConfig
    recordChannelCacheResult
else
    echo "No need to populate channel fixtures"
fi