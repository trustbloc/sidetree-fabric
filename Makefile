#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
#   all (default) : runs code checks and unit tests
#   checks: runs code checks (license, spelling, lint)
#   unit-test: runs unit tests


GO_CMD ?= go
export GO111MODULE=on

checks: license lint

license:
	@scripts/check_license.sh

lint:
	@scripts/check_lint.sh

unit-test:
	@scripts/unit.sh

all: checks unit-test





