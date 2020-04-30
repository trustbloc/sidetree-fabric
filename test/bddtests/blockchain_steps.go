/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"

	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

// BlockchainSteps
type BlockchainSteps struct {
	resp       *restclient.HttpResponse
	bddContext *bddtests.BDDContext
}

// NewBlockchainSteps
func NewBlockchainSteps(context *bddtests.BDDContext) *BlockchainSteps {
	return &BlockchainSteps{bddContext: context}
}

func (d *BlockchainSteps) hashOfBase64EncodedValueEquals(base64EncodedValue, base64EncodedHash string) error {
	var err error

	base64EncodedHash, err = bddtests.ResolveVarsInExpression(base64EncodedHash)
	if err != nil {
		return err
	}

	base64EncodedValue, err = bddtests.ResolveVarsInExpression(base64EncodedValue)
	if err != nil {
		return err
	}

	hash, err := base64.StdEncoding.DecodeString(base64EncodedHash)
	if err != nil {
		return errors.WithMessagef(err, "hash is not base64 URL-encoded")
	}

	data, err := base64.StdEncoding.DecodeString(base64EncodedValue)
	if err != nil {
		return errors.WithMessagef(err, "value is not base64-encoded")
	}

	sum := sha256.Sum256(data)

	if !bytes.Equal(sum[:], hash) {
		return errors.Errorf("The provided hash does not match the hash of the value")
	}

	logger.Infof("The hash of the value matches the provided hash")

	return nil
}

func (d *BlockchainSteps) hashOfBase64URLEncodedValueEquals(base64URLEncodedValue, base64URLEncodedHash string) error {
	var err error

	base64URLEncodedHash, err = bddtests.ResolveVarsInExpression(base64URLEncodedHash)
	if err != nil {
		return err
	}

	base64URLEncodedValue, err = bddtests.ResolveVarsInExpression(base64URLEncodedValue)
	if err != nil {
		return err
	}

	hash, err := base64.URLEncoding.DecodeString(base64URLEncodedHash)
	if err != nil {
		return errors.WithMessagef(err, "hash is not base64 URL-encoded")
	}

	data, err := base64.URLEncoding.DecodeString(base64URLEncodedValue)
	if err != nil {
		return errors.WithMessagef(err, "value is not base64 URL-encoded")
	}

	sum := sha256.Sum256(data)

	if !bytes.Equal(sum[:], hash) {
		return errors.Errorf("The provided hash does not match the hash of the value")
	}

	logger.Infof("The hash of the value matches the provided hash")

	return nil
}

// RegisterSteps registers did sidetree steps
func (d *BlockchainSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^the hash of the base64-encoded value "([^"]*)" equals "([^"]*)"$`, d.hashOfBase64EncodedValueEquals)
	s.Step(`^the hash of the base64URL-encoded value "([^"]*)" equals "([^"]*)"$`, d.hashOfBase64URLEncodedValueEquals)
}