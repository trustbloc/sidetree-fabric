/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationqueue

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-fabric/pkg/context/operationqueue/mocks"
)

//go:generate counterfeiter -o ./mocks/dbhandle.gen.go --fake-name DBHandle . dbHandle

const (
	channel1 = "channel1"
	channel2 = "channel2"
	channel3 = "channel3"

	namespace1 = "namespace1"
)

var (
	op1 = &operation.QueuedOperation{UniqueSuffix: "op1"}
	op2 = &operation.QueuedOperation{UniqueSuffix: "op2"}
	op3 = &operation.QueuedOperation{UniqueSuffix: "op3"}
	op4 = &operation.QueuedOperation{UniqueSuffix: "op4"}
)

func TestLevelDBQueue(t *testing.T) {
	q, cleanup, err := newTestQueue(channel1)
	require.NoError(t, err)

	defer cleanup()

	require.Zero(t, q.Len())

	ops, err := q.Peek(2)
	require.NoError(t, err)
	require.Empty(t, ops)

	n, err := q.Add(op1, 100)
	require.NoError(t, err)
	require.Equal(t, uint(1), n)

	n, err = q.Add(op2, 101)
	require.NoError(t, err)
	require.Equal(t, uint(2), n)

	n, err = q.Add(op3, 101)
	require.NoError(t, err)
	require.Equal(t, uint(3), n)

	ops, err = q.Peek(2)
	require.NoError(t, err)
	require.Len(t, ops, 2)
	require.Equal(t, *op1, ops[0].QueuedOperation)
	require.Equal(t, uint64(100), ops[0].ProtocolGenesisTime)
	require.Equal(t, *op2, ops[1].QueuedOperation)
	require.Equal(t, uint64(101), ops[1].ProtocolGenesisTime)

	removed, n, err := q.Remove(2)
	require.NoError(t, err)
	require.Equal(t, uint(1), n)
	require.Equal(t, uint(2), removed)
	require.Equal(t, uint(1), q.Len())

	removed, n, err = q.Remove(2)
	require.NoError(t, err)
	require.Equal(t, uint(0), n)
	require.Equal(t, uint(1), removed)
}

func TestLevelDBQueue_Reload(t *testing.T) {
	q, cleanup, err := newTestQueue(channel2)
	require.NoError(t, err)

	defer cleanup()
	require.Zero(t, q.Len())

	n, err := q.Add(op1, 100)
	require.NoError(t, err)
	require.Equal(t, uint(1), n)

	n, err = q.Add(op2, 100)
	require.NoError(t, err)
	require.Equal(t, uint(2), n)

	n, err = q.Add(op3, 101)
	require.NoError(t, err)
	require.Equal(t, uint(3), n)

	n, err = q.Add(op4, 101)
	require.NoError(t, err)
	require.Equal(t, uint(4), n)
	require.Equal(t, uint(4), q.Len())

	removed, l, err := q.Remove(1)
	require.NoError(t, err)
	require.Equal(t, uint(1), removed)

	require.Equal(t, uint(3), l)
	require.Equal(t, uint(3), q.Len())

	q.Close()

	q2, cleanup2, err := newTestQueue(channel2)
	require.NoError(t, err)
	defer cleanup2()

	require.Equal(t, uint(3), q2.Len())

	removed, l, err = q2.Remove(1)
	require.NoError(t, err)
	require.Equal(t, uint(2), l)
	require.Equal(t, uint(1), removed)

	removed, l, err = q2.Remove(2)
	require.NoError(t, err)
	require.Zero(t, l)
	require.Equal(t, uint(2), removed)

	q2.Close()

	q3, cleanup3, err := newTestQueue(channel2)
	require.NoError(t, err)
	defer cleanup3()
	require.Zero(t, q3.Len())
}

func TestLevelDBQueue_Close(t *testing.T) {
	q, err := newLevelDBQueue(channel3, namespace1, levelDBBasePath)
	require.NoError(t, err)
	defer func() {
		if err := q.Drop(); err != nil {
			t.Errorf("Error dropping DB [%s]: %s", q.dir, err)
		}
	}()

	err = q.Drop()
	require.Error(t, err, errNotClosed.Error())

	q.Close()
	require.NotPanicsf(t, func() { q.Close() }, "calling close twice should not panic")

	_, err = q.Add(&operation.QueuedOperation{}, 100)
	require.EqualError(t, err, errClosed.Error())

	_, err = q.Peek(1)
	require.EqualError(t, err, errClosed.Error())

	_, _, err = q.Remove(1)
	require.EqualError(t, err, errClosed.Error())

	require.Zero(t, q.Len())
}

func TestLevelDBQueue_Error(t *testing.T) {
	t.Run("Create DB error", func(t *testing.T) {
		errExpected := errors.New("injected OpenFile error")

		openFile = func(dir string) (handle dbHandle, err error) {
			return nil, errExpected
		}

		q, err := newLevelDBQueue(channel1, namespace1, levelDBBasePath)
		require.Error(t, err, errExpected.Error())
		require.Nil(t, q)
	})

	t.Run("Add error", func(t *testing.T) {
		errExpected := errors.New("injected Add error")

		openFile = func(dir string) (handle dbHandle, err error) {
			db := &mocks.DBHandle{}
			db.NewIteratorReturns(&mocks.Iterator{})
			db.PutReturns(errExpected)

			return db, nil
		}

		q, err := newLevelDBQueue(channel1, namespace1, levelDBBasePath)
		require.NoError(t, err)
		require.NotNil(t, q)

		_, err = q.Add(&operation.QueuedOperation{}, 100)
		require.Error(t, err, errExpected.Error())
	})

	t.Run("Remove error", func(t *testing.T) {
		errExpected := errors.New("injected Remove error")

		openFile = func(dir string) (handle dbHandle, err error) {
			it := &mocks.Iterator{}
			it.NextReturnsOnCall(0, true)
			it.KeyReturns(toBytes(1000))

			v, err := marshal(&operation.QueuedOperationAtTime{})
			require.NoError(t, err)
			it.ValueReturns(v)

			db := &mocks.DBHandle{}
			db.NewIteratorReturns(it)
			db.DeleteReturns(errExpected)

			return db, nil
		}

		q, err := newLevelDBQueue(channel1, namespace1, levelDBBasePath)
		require.NoError(t, err)
		require.NotNil(t, q)

		_, err = q.Add(&operation.QueuedOperation{}, 100)
		require.NoError(t, err)

		_, _, err = q.Remove(1)
		require.Error(t, err, errExpected.Error())
	})
}

func newTestQueue(channelID string) (q *LevelDBQueue, cleanup func(), err error) {
	q, err = newLevelDBQueue(channelID, namespace1, levelDBBasePath)
	if err != nil {
		return nil, nil, err
	}

	cleanup = func() {
		q.Close()

		if err := q.Drop(); err != nil {
			logger.Warnf("Error dropping queue: %v", err)
		}
	}

	return
}
