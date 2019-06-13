#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	gcc \
	musl-dev \
	git \
	libtool \
	bash \
	make;
ADD . $GOPATH/src/github.com/trustbloc/sidetree-fabric
WORKDIR $GOPATH/src/github.com/trustbloc/sidetree-fabric
ENV EXECUTABLES go git

FROM golang as sidetree
ARG GO_TAGS
ARG GOPROXY
RUN GO_TAGS=${GO_TAGS} GOPROXY=${GOPROXY} make sidetree


FROM alpine:${ALPINE_VER} as base
COPY --from=sidetree /go/src/github.com/trustbloc/sidetree-fabric/.build/bin/sidetree-fabric /usr/local/bin
CMD ["sidetree-fabric"]