/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
)

type blockIterator struct {
	currentBlockNum uint64
	sinceTxnNum     uint64
	height          uint64
}

type blockDesc struct {
	blockNum uint64
	txnNum   uint64
}

func (d *blockDesc) BlockNum() uint64 {
	return d.blockNum
}

func (d *blockDesc) TxnNum() uint64 {
	return d.txnNum
}

func newBlockIterator(bcInfo *cb.BlockchainInfo, sinceBlockNum, sinceTxnNum uint64) *blockIterator {
	return &blockIterator{
		currentBlockNum: sinceBlockNum,
		sinceTxnNum:     sinceTxnNum,
		height:          bcInfo.Height,
	}
}

func (it *blockIterator) Next() (BlockDescriptor, bool) {
	if it.currentBlockNum == it.height {
		return nil, false
	}

	desc := &blockDesc{
		blockNum: it.currentBlockNum,
		txnNum:   it.sinceTxnNum,
	}

	// Reset sinceTxnNum to 0 since it only applies to the first block
	it.sinceTxnNum = 0
	it.currentBlockNum++

	return desc, true
}
