#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#


# Release Parameters
BASE_VERSION=0.1.5
IS_RELEASE=false

# Project Parameters
SOURCE_REPO=sidetree-fabric
BASE_PKG_NAME=peer
RELEASE_REPO=docker.pkg.github.com/trustbloc/${SOURCE_REPO}
SNAPSHOT_REPO=docker.pkg.github.com/trustbloc-cicd/snapshot

ARCH=$(go env GOARCH)

if [ ${IS_RELEASE} = false ]
then
  EXTRA_VERSION=snapshot-$(git rev-parse --short=7 HEAD)
  PROJECT_VERSION=${BASE_VERSION}-${EXTRA_VERSION}
  PROJECT_PKG_REPO=${SNAPSHOT_REPO}
else
  PROJECT_VERSION=${BASE_VERSION}
  PROJECT_PKG_REPO=${RELEASE_REPO}
fi

export SIDETREE_FABRIC_TAG=$ARCH-$PROJECT_VERSION
export SIDETREE_FABRIC_PKG=${PROJECT_PKG_REPO}/${BASE_PKG_NAME}
