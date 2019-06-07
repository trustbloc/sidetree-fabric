/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	client, err := New("testdata/protocol.json")
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestNewError(t *testing.T) {
	client, err := New("testdata/invalid.json")
	assert.NotNil(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestCurrentProtocol(t *testing.T) {
	client, err := New("testdata/protocol.json")
	assert.Nil(t, err)
	assert.NotNil(t, client)

	protocol := client.Current()
	assert.Equal(t, uint(10000), protocol.MaxOperationsPerBatch)
}
