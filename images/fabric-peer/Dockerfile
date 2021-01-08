# Copyright SecureKey Technologies Inc. and the TrustBloc contributors. All Rights Reserved.
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Original image template from: https://github.com/hyperledger/fabric/tree/master/images/peer

ARG GO_VER
ARG ALPINE_VER
ARG FABRIC_PEER_UPSTREAM_IMAGE
ARG FABRIC_PEER_UPSTREAM_TAG

FROM alpine:${ALPINE_VER} as peer-base
RUN apk add --no-cache tzdata

FROM ${FABRIC_PEER_UPSTREAM_IMAGE}:${FABRIC_PEER_UPSTREAM_TAG} as peer-upstream

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	bash \
	gcc \
        binutils-gold \
	git \
	make \
	musl-dev
ADD . $GOPATH/src/github.com/trustbloc/sidetree-fabric
WORKDIR $GOPATH/src/github.com/trustbloc/sidetree-fabric

FROM golang as peer
ARG GO_TAGS
ARG GO_LDFLAGS
RUN make GO_TAGS="${GO_TAGS}" GO_LDFLAGS="${GO_LDFLAGS}" fabric-peer

FROM peer-base
LABEL org.opencontainers.image.source https://github.com/trustbloc/sidetree-fabric
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric
VOLUME /etc/hyperledger/fabric
VOLUME /var/hyperledger
COPY --from=peer /go/src/github.com/trustbloc/sidetree-fabric/.build/bin /usr/local/bin
COPY --from=peer-upstream ${FABRIC_CFG_PATH} ${FABRIC_CFG_PATH}
EXPOSE 7051
CMD ["fabric-peer"]
