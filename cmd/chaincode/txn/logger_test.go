/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"testing"
)

func TestLogger(t *testing.T) {
	l := NewLogger("test")
	arg := "some arg"
	l.Debugf("Some message: [%s]", arg)
	l.Infof("Some message: [%s]", arg)
	l.Warnf("Some message: [%s]", arg)
	l.Errorf("Some message: [%s]", arg)
}
