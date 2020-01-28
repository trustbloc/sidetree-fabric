/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	fabricCLIPath = "../../.build/bin/fabric"
	homeDir       = "./.fabriccli/"
)

// FabricCLI invokes the fabric-cli command-line tool
type FabricCLI struct {
}

// NewFabricCLI returns a new NewFabricCLI
func NewFabricCLI() *FabricCLI {
	return &FabricCLI{}
}

// Exec executes fabric-cli with the given args and returns the response
func (cli *FabricCLI) Exec(args ...string) (string, error) {
	var a []string
	a = append(a, "--home", homeDir)
	a = append(a, args...)
	cmd := exec.Command(fabricCLIPath, a...)
	cmd.Env = []string{"PROJECT_PATH=../.."}

	var out bytes.Buffer
	var er bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &er

	err := cmd.Start()
	if err != nil {
		return er.String(), errors.Errorf("%s: %s", err, er.Bytes())
	}
	err = cmd.Wait()
	if err != nil {
		return er.String(), errors.Errorf("%s: %s", err, er.Bytes())
	}
	return out.String(), nil
}
