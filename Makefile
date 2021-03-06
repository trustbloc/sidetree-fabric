# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
#   all:                        runs code checks and unit tests
#   clean:                      cleans the build area
#   checks:                     runs code checks (license, spelling, lint)
#   unit-test:                  runs unit tests
#   populate-fixtures:          populate crypto directory and channel configuration for bddtests
#   crypto-gen:                 generates crypto directory
#   channel-config-gen:         generates test channel configuration transactions and blocks
#   bddtests:                   run bddtests
#   docker-thirdparty:          pulls thirdparty images
#   generate:                   generates mocks
#


# Tool commands (overridable)
DOCKER_CMD ?= docker
GO_CMD     ?= go
ALPINE_VER ?= 3.12
DBUILD_CMD ?= docker build

# defined in github.com/hyperledger/fabric/common/metadata/metadata.go
FABRIC_METADATA_VAR = Version=2.2.1

# Build flags
GO_TAGS    ?=
GO_LDFLAGS ?= $(FABRIC_METADATA_VAR:%=-X github.com/hyperledger/fabric/common/metadata.%)

# Local variables used by makefile
PROJECT_NAME       = sidetree-fabric
CONTAINER_IDS      = $(shell docker ps -a -q)
DEV_IMAGES         = $(shell docker images dev-* -q)
ARCH               = $(shell go env GOARCH)
GO_VER             = 1.14.4
export GO111MODULE = on

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE   ?= hyperledger/fabric-tools
# TODO: fabric-sdk-go fails when using artifacts generated by fabric-tools v2.0.0. Switch to fabric-tools v2.0.0 after the SDK is fixed.
# FABRIC_TOOLS_VERSION ?= 2.0.0
FABRIC_TOOLS_VERSION ?= 2.0.0-alpha
FABRIC_TOOLS_TAG     ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric peer ext docker image (overridable)
FABRIC_PEER_IMAGE   ?= hyperledger/fabric-peer
FABRIC_PEER_VERSION ?= 2.2.1
FABRIC_PEER_TAG     ?= $(ARCH)-$(FABRIC_PEER_VERSION)

export FABRIC_CLI_EXT_VERSION ?= v0.1.5

# Namespace for the Sidetree Fabric peer image
DOCKER_OUTPUT_NS           ?= ghcr.io/trustbloc
SIDETREE_FABRIC_IMAGE_NAME ?= sidetree-fabric
SIDETREE_FABRIC_IMAGE_TAG  ?= latest

checks: license lint

.PHONY: license
license:
	@scripts/check_license.sh

lint:
	@scripts/check_lint.sh

unit-test: checks
	@scripts/unit.sh

generate:
	go generate ./...

all: clean checks unit-test bddtests


crypto-gen:
	@echo "Generating crypto directory ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric /opt/workspace/${PROJECT_NAME}/scripts/generate_crypto.sh"

channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/ /opt/workspace/${PROJECT_NAME}/scripts/generate_channeltx.sh"

populate-fixtures:
	@scripts/populate-fixtures.sh -f


bddtests: clean populate-fixtures docker-thirdparty fabric-peer-docker build-cc fabric-cli
	@scripts/integration.sh


fabric-peer:
	@echo "Building fabric-peer"
	@mkdir -p ./.build/bin
	@cd cmd/peer && go build -tags "$(GO_TAGS)" -ldflags "$(GO_LDFLAGS)" \
        -o ../../.build/bin/fabric-peer main.go

fabric-peer-docker:
	@echo "Building fabric-peer image"
	@$(DBUILD_CMD) \
        -f ./images/fabric-peer/Dockerfile --no-cache \
        -t $(DOCKER_OUTPUT_NS)/$(SIDETREE_FABRIC_IMAGE_NAME):$(SIDETREE_FABRIC_IMAGE_TAG) \
	--build-arg FABRIC_PEER_UPSTREAM_IMAGE=$(FABRIC_PEER_IMAGE) \
	--build-arg FABRIC_PEER_UPSTREAM_TAG=$(FABRIC_PEER_TAG) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg GO_LDFLAGS="$(GO_LDFLAGS)" \
	--build-arg GO_TAGS="$(GO_TAGS)" .

docker-thirdparty:
	docker pull couchdb:3.1
	docker pull hyperledger/fabric-orderer:$(ARCH)-2.2.0

build-cc:
	@echo "Building cc"
	@mkdir -p ./.build
	@scripts/copycc.sh

fabric-cli:
	@scripts/build_fabric_cli.sh

clean-images:
	@echo "Stopping all containers, pruning containers and images, deleting dev images"
ifneq ($(strip $(CONTAINER_IDS)),)
	@docker stop $(CONTAINER_IDS)
endif
	@docker system prune -f
ifneq ($(strip $(DEV_IMAGES)),)
	@docker rmi $(DEV_IMAGES) -f
endif

clean:
	rm -Rf ./test/bddtests/docker-compose.log
	rm -Rf ./test/bddtests/fixtures/fabric/channel
	rm -Rf ./test/bddtests/fixtures/fabric/crypto-config
	rm -Rf ./.build
	rm -Rf ./test/bddtests/.fabriccli
