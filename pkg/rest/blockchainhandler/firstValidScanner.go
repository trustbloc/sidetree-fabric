/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"bytes"
	"encoding/base64"
	"strings"

	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var errInvalidTransaction = errors.New("invalid transaction")
var errInvalidDesc = errors.New("invalid descriptor")
var errFoundValid = errors.New("found valid transaction")

type firstValidScanner struct {
	channelID string
	bcClient  bcclient.Blockchain
}

func newFirstValidScanner(channelID string, bcClient bcclient.Blockchain) *firstValidScanner {
	return &firstValidScanner{
		channelID: channelID,
		bcClient:  bcClient,
	}
}

type firstValidBlockDescriptor interface {
	BlockDescriptor
	Transaction() Transaction
}

// Scan scans the given block in order to find out if the transaction is valid.
// The block descriptor must be of type *firstValidDesc.
func (p *firstValidScanner) Scan(desc BlockDescriptor, _ int) ([]Transaction, bool, error) {
	fvDesc, ok := desc.(firstValidBlockDescriptor)
	if !ok {
		return nil, false, errInvalidDesc
	}

	isValid, err := newTxnValidator(p.channelID, fvDesc, p.bcClient).isValid()
	if err != nil {
		return nil, false, err
	}

	if isValid {
		return []Transaction{fvDesc.Transaction()}, true, nil
	}

	return nil, false, nil
}

type txnValidator struct {
	channelID string
	desc      firstValidBlockDescriptor
	bcClient  bcclient.Blockchain
}

func newTxnValidator(channelID string, desc firstValidBlockDescriptor, bcClient bcclient.Blockchain) *txnValidator {
	return &txnValidator{
		channelID: channelID,
		desc:      desc,
		bcClient:  bcClient,
	}
}

func (v *txnValidator) isValid() (bool, error) {
	block, err := v.bcClient.GetBlockByNumber(v.desc.BlockNum())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debugf("[%s] Block [%d] not found", v.channelID, v.desc.BlockNum())

			return false, nil
		}

		return false, err
	}

	blockHash := protoutil.BlockHeaderHash(block.Header)
	txnHash, err := base64.URLEncoding.DecodeString(v.desc.Transaction().TransactionTimeHash)
	if err != nil {
		logger.Debugf("[%s] Invalid base64 encoded transaction_time_hash for transaction_time: %s", v.channelID, v.desc.BlockNum(), err)

		return false, nil
	}

	if !bytes.Equal(blockHash, txnHash) {
		logger.Debugf("[%s] transaction_time_hash does not match the hash of the block header for block [%d]", v.channelID, v.desc.BlockNum())

		return false, nil
	}

	visitor := blockvisitor.New(v.channelID,
		blockvisitor.WithWriteHandler(v.handleWrite),
		blockvisitor.WithErrorHandler(v.handleError),
	)

	err = visitor.Visit(block)
	if err != nil {
		if errors.Cause(err) == errFoundValid {
			return true, nil
		}

		if errors.Cause(err) == errInvalidTransaction {
			return false, nil
		}

		return false, err
	}

	logger.Debugf("[%s] TransactionNumber [%d] not found in block [%d]", v.channelID, v.desc.TxnNum(), v.desc.BlockNum())

	return false, nil
}

func (v *txnValidator) handleWrite(w *blockvisitor.Write) error {
	if !strings.HasPrefix(w.Write.Key, common.AnchorPrefix) {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", v.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorPrefix)

		return nil
	}

	if w.TxNum != v.desc.TxnNum() {
		logger.Debugf("[%s] Ignoring write in block [%d] and TxNum [%d] since it doesn't match transaction number %d", v.channelID, w.BlockNum, w.TxNum, v.desc.TxnNum())

		return nil
	}

	anchorString, err := getAnchorString(w.Write.Value)
	if err != nil {
		return errors.WithMessagef(err, "failed to get anchor string [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
	}

	if anchorString != v.desc.Transaction().AnchorString {
		logger.Debugf("[%s] AnchorString [%s] for block [%d] and TxnNumber [%d] does not match the provided AnchorString [%s]", v.channelID, anchorString, v.desc.BlockNum(), v.desc.TxnNum(), v.desc.Transaction().AnchorString)

		return errInvalidTransaction
	}

	// Found a valid transaction - stop processing
	return errFoundValid
}

func (v *txnValidator) handleError(err error, ctx *blockvisitor.Context) error {
	if err == errInvalidTransaction {
		logger.Debugf("[%s] Got invalid transaction %+v", v.channelID, v.desc.Transaction)

		return err
	}

	if err == errFoundValid {
		logger.Debugf("[%s] Found valid transaction %+v", v.channelID, v.desc.Transaction)

		return err
	}

	logger.Errorf("[%s] Error processing block: %s. Context: %s. Block will be ignored.", v.channelID, err, ctx)

	return nil
}
