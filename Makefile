#
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
#   sidetree-docker:            build sidetree-fabric image
#   generate:                   generates mocks
#


# Tool commands (overridable)
DOCKER_CMD ?= docker
GO_CMD     ?= go
ALPINE_VER ?= 3.9
GO_TAGS    ?=

# Local variables used by makefile
PROJECT_NAME       = sidetree-fabric
CONTAINER_IDS      = $(shell docker ps -a -q)
DEV_IMAGES         = $(shell docker images dev-* -q)
ARCH               = $(shell go env GOARCH)
GO_VER             = $(shell grep "GO_VER" .ci-properties |cut -d'=' -f2-)
export GO111MODULE = on

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE   ?= hyperledger/fabric-tools
FABRIC_TOOLS_VERSION ?= 2.0.0-alpha
FABRIC_TOOLS_TAG     ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric peer ext docker image (overridable)
FABRIC_PEER_EXT_IMAGE   ?= trustbloc/fabric-peer
FABRIC_PEER_EXT_VERSION ?= 0.1.0-snapshot-1ab7e44
FABRIC_PEER_EXT_TAG     ?= $(ARCH)-$(FABRIC_PEER_EXT_VERSION)

# Namespace for the blocnode image
DOCKER_OUTPUT_NS     ?= trustbloc
SIDETREE_FABRIC_IMAGE_NAME ?= sidetree-fabric

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


bddtests: clean checks populate-fixtures docker-thirdparty bddtests-fabric-peer-docker sidetree-docker build-cc
	@scripts/integration.sh


bddtests-fabric-peer-cli:
	@echo "Building fabric-peer cli"
	@mkdir -p ./.build/bin
	@cd test/bddtests/fixtures/fabric/peer/cmd && go build -o ../../../../../../.build/bin/fabric-peer github.com/trustbloc/sidetree-fabric/test/bddtests/fixtures/fabric/peer/cmd

bddtests-fabric-peer-docker:
	@docker build -f ./test/bddtests/fixtures/images/fabric-peer/Dockerfile --no-cache -t fabric-peer:latest \
	--build-arg FABRIC_PEER_EXT_IMAGE=$(FABRIC_PEER_EXT_IMAGE) \
	--build-arg FABRIC_PEER_EXT_TAG=$(FABRIC_PEER_EXT_TAG) \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

docker-thirdparty:
	docker pull couchdb:2.2.0
	docker pull hyperledger/fabric-orderer:$(ARCH)-2.0.0-alpha

sidetree:
	@echo "Building sidetree"
	@mkdir -p ./.build/bin
	@go build -o ./.build/bin/sidetree-fabric cmd/sidetree-server/main.go

sidetree-docker:
	@docker build -f ./images/sidetree-fabric/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(SIDETREE_FABRIC_IMAGE_NAME):latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .


build-cc:
	@echo "Building cc"
	@mkdir -p ./.build
	@scripts/copycc.sh

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





