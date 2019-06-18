/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/json"
	"fmt"

	"github.com/DATA-DOG/godog"

	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

// SidetreeSteps ...
type SidetreeSteps struct {
	BDDContext *bddtests.BDDContext
	content    string
	address    string
}

// NewSidetreeSteps define custom steps
func NewSidetreeSteps(context *bddtests.BDDContext) *SidetreeSteps {
	return &SidetreeSteps{BDDContext: context}
}

func (t *SidetreeSteps) writeContent(content, ccID, orgIDs, channelID string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	args := []string{"writeContent", content}
	resp, err := commonSteps.InvokeCCWithArgs(ccID, channelID, commonSteps.OrgPeers(orgIDs, channelID), args, nil)
	if err != nil {
		return fmt.Errorf("InvokeCCWithArgs return error: %s", err)
	}

	t.content = content
	t.address = string(resp.Payload)

	return nil
}

func (t *SidetreeSteps) readContent(ccID, orgIDs, channelID string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	args := []string{"readContent", t.address}
	payload, err := commonSteps.QueryCCWithArgs(false, ccID, channelID, args, nil, commonSteps.OrgPeers(orgIDs, channelID)...)
	if err != nil {
		return fmt.Errorf("QueryCCWithArgs return error: %s", err)
	}

	if payload != t.content {
		return fmt.Errorf("original content[%s] doesn't match retrieved content[%s]", t.content, payload)
	}

	return nil
}

func (t *SidetreeSteps) anchorBatch(didID, ccID, orgIDs, channelID string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	// Create default operations for this did (for now two create and update)
	operations := getDefaultOperations(didID)

	batchFile := getBatchFileBytes(operations)
	err := t.writeContent(batchFile, ccID, orgIDs, channelID)
	if err != nil {
		return fmt.Errorf("write batch file to DCAS return error: %s", err)
	}

	anchor := getAnchorFileBytes(t.address, "")
	err = t.writeContent(anchor, ccID, orgIDs, channelID)
	if err != nil {
		return fmt.Errorf("write anchor file to DCAS return error: %s", err)
	}

	args := []string{"writeAnchor", t.address}
	_, err = commonSteps.InvokeCCWithArgs(ccID, channelID, commonSteps.OrgPeers(orgIDs, channelID), args, nil)
	if err != nil {
		return fmt.Errorf("InvokeCCWithArgs return error: %s", err)
	}

	return nil
}

func getJSON(op Operation) string {

	bytes, err := json.Marshal(op)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// Operation defines sample operation
type Operation struct {
	//Operation type
	Type string
	//ID is full ID for this document - includes namespace + unique suffix
	ID string
}

// AnchorFile defines the schema of a Anchor File and its related operations.
type AnchorFile struct {
	// BatchFileHash is encoded hash of the batch file
	BatchFileHash string `json:"batchFileHash"`

	// MerkleRoot is encoded root hash of the Merkle tree constructed from
	// the operations included in the batch file
	MerkleRoot string `json:"merkleRoot"`
}

// BatchFile defines the schema of a Batch File and its related operations.
type BatchFile struct {
	// operations included in this batch file, each operation is an encoded string
	Operations []string `json:"operations"`
}

func getCreateOperation(did string) string {
	op := Operation{ID: did, Type: "create"}
	return getJSON(op)
}

func getUpdateOperation(did string) string {
	op := Operation{ID: did, Type: "update"}
	return getJSON(op)
}

func getDefaultOperations(did string) []string {
	return []string{getCreateOperation(did), getUpdateOperation(did)}
}

func getBatchFileBytes(operations []string) string {
	bf := BatchFile{Operations: operations}
	bytes, err := json.Marshal(bf)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func getAnchorFileBytes(batchFileHash string, merkleRoot string) string {
	af := AnchorFile{
		BatchFileHash: batchFileHash,
		MerkleRoot:    merkleRoot,
	}
	s, err := json.Marshal(af)
	if err != nil {
		panic(err)
	}
	return string(s)
}

// RegisterSteps registers sidetree txn steps
func (t *SidetreeSteps) RegisterSteps(s *godog.Suite) {
	s.BeforeScenario(t.BDDContext.BeforeScenario)
	s.AfterScenario(t.BDDContext.AfterScenario)
	s.Step(`^client writes content "([^"]*)" using "([^"]*)" on all peers in the "([^"]*)" org on the "([^"]*)" channel$`, t.writeContent)
	s.Step(`^client verifies that written content at the returned address from "([^"]*)" matches original content on all peers in the "([^"]*)" org on the "([^"]*)" channel$`, t.readContent)
	s.Step(`^client writes operations batch file and anchor file for ID "([^"]*)" using "([^"]*)" on all peers in the "([^"]*)" org on the "([^"]*)" channel$`, t.anchorBatch)
}
