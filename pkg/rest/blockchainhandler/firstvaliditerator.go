/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

type firstValidDesc struct {
	txn Transaction
}

func (d *firstValidDesc) BlockNum() uint64 {
	return d.txn.TransactionTime
}

func (d *firstValidDesc) TxnNum() uint64 {
	return d.txn.TransactionNumber
}

func (d *firstValidDesc) Transaction() Transaction {
	return d.txn
}

type firstValidIterator struct {
	current      int
	transactions []Transaction
}

func newFirstValidIterator(transactions []Transaction) *firstValidIterator {
	return &firstValidIterator{
		transactions: transactions,
	}
}

func (it *firstValidIterator) Next() (BlockDescriptor, bool) {
	if it.current >= len(it.transactions) {
		return nil, false
	}

	desc := &firstValidDesc{
		txn: it.transactions[it.current],
	}

	it.current++

	return desc, true
}
