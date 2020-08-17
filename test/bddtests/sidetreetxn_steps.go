/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

// SidetreeTxnSteps ...
type SidetreeTxnSteps struct {
	BDDContext *bddtests.BDDContext
	content    string
	address    string
}

// NewSidetreeSteps define custom steps
func NewSidetreeSteps(context *bddtests.BDDContext) *SidetreeTxnSteps {
	return &SidetreeTxnSteps{BDDContext: context}
}

func (t *SidetreeTxnSteps) writeContent(content, ccID, coll, channelID string) error {
	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	args := []string{"writeContent", coll, content}
	resp, err := commonSteps.QueryCCWithArgs(false, ccID, channelID, "", args, nil)
	if err != nil {
		return fmt.Errorf("QueryCCWithArgs return error: %s", err)
	}

	t.content = content
	t.address = resp

	return nil
}

func (t *SidetreeTxnSteps) readContent(ccID, coll, channelID string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	args := []string{"readContent", coll, t.address}
	payload, err := commonSteps.QueryCCWithArgs(false, ccID, channelID, "", args, nil)
	if err != nil {
		return fmt.Errorf("QueryCCWithArgs return error: %s", err)
	}

	if payload != t.content {
		return fmt.Errorf("original content[%s] doesn't match retrieved content[%s]", t.content, payload)
	}

	return nil
}

func (t *SidetreeTxnSteps) anchorBatch(didID, ccID, coll, channelID string) error {
	logger.Infof("Preparing to write anchor batch on channel [%s]", channelID)
	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	// Create default encoded operations for this did (for now two create and update)
	operations := getDefaultOperations(didID)

	batchFile := getBatchFileBytes(operations)
	logger.Infof("... writing batch file on channel [%s]", channelID)
	err := t.writeContent(batchFile, ccID, coll, channelID)
	if err != nil {
		return fmt.Errorf("write batch file to DCAS return error: %s", err)
	}

	anchor := getAnchorFileBytes(t.address, []string{"uniqueSuffix1", "uniqueSuffix2"})
	logger.Infof("... writing anchor file on channel [%s]", channelID)
	err = t.writeContent(anchor, ccID, coll, channelID)
	if err != nil {
		return fmt.Errorf("write anchor file to DCAS return error: %s", err)
	}

	logger.Infof("... committing anchor address [%s] to ledger on channel [%s]", t.address, channelID)
	args := []string{"writeAnchor", t.address}
	_, err = commonSteps.InvokeCCWithArgs(ccID, channelID, "", nil, args, nil)
	if err != nil {
		return fmt.Errorf("InvokeCCWithArgs return error: %s", err)
	}

	return nil
}

func (t *SidetreeTxnSteps) writeDocument(op string, ccID, channelID string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	args := []string{"write", op}
	_, err := commonSteps.InvokeCCWithArgs(ccID, channelID, "", nil, args, nil)
	if err != nil {
		return fmt.Errorf("InvokeCCWithArgs return error: %s", err)
	}

	return nil
}

func (t *SidetreeTxnSteps) createDocument(docID, ccID, channelID string) error {
	return t.writeDocument(getCreateOperation(docID), ccID, channelID)
}

func (t *SidetreeTxnSteps) updateDocument(docID, ccID, channelID string) error {
	return t.writeDocument(getUpdateOperation(docID), ccID, channelID)
}

func (t *SidetreeTxnSteps) queryDocumentByIndex(docID, ccID, numOfDocs, channelID string) error {
	return t.queryDocumentByIndexOnTargets(docID, ccID, numOfDocs, channelID, "")
}

func (t *SidetreeTxnSteps) queryDocumentByIndexOnTargets(docID, ccID, numOfDocs, channelID, peerIDs string) error {

	commonSteps := bddtests.NewCommonSteps(t.BDDContext)

	var targets bddtests.Peers
	if peerIDs != "" {
		logger.Infof("Querying for document [%s] on peers [%s]", docID, peerIDs)
		var err error
		targets, err = commonSteps.Peers(peerIDs)
		if err != nil {
			return err
		}
	} else {
		logger.Infof("Querying for document [%s]", docID)
	}

	args := []string{"queryByID", docID}
	payload, err := commonSteps.QueryCCWithArgs(false, ccID, channelID, "", args, nil, targets...)
	if err != nil {
		return fmt.Errorf("QueryCCWithArgs return error: %s", err)
	}

	var operations [][]byte
	err = json.Unmarshal([]byte(payload), &operations)
	if err != nil {
		return fmt.Errorf("failed to unmarshal operations: %s", err)
	}

	docsNum, err := strconv.Atoi(numOfDocs)
	if err != nil {
		return err
	}

	if len(operations) != docsNum {
		return fmt.Errorf("expecting %d, got %d", docsNum, len(operations))
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
	Type string `json:"type"`
	//ID is full ID for this document - includes namespace + unique suffix
	ID string `json:"id"`
}

// AnchorFile defines the schema of a Anchor File and its related operations.
type AnchorFile struct {
	// BatchFileHash is encoded hash of the batch file
	BatchFileHash string `json:"batchFileHash"`

	// UniqueSuffixes is an array of suffixes (the unique portion of the ID string that differentiates
	// one document from another) for all documents that are declared to have operations within the associated batch file.
	UniqueSuffixes []string `json:"uniqueSuffixes"`
}

// BatchFile defines the schema of a Batch File and its related operations.
type BatchFile struct {
	// operations included in this batch file, each operation is an encoded string
	Operations []string `json:"operations"`
}

func getCreateOperation(id string) string {
	op := Operation{ID: id, Type: "create"}
	return getJSON(op)
}

func getUpdateOperation(id string) string {
	op := Operation{ID: id, Type: "update"}
	return getJSON(op)
}

func encode(op string) string {
	return base64.URLEncoding.EncodeToString([]byte(op))
}

func getDefaultOperations(id string) []string {
	return []string{encode(getCreateOperation(id)), encode(getUpdateOperation(id))}
}

func getBatchFileBytes(operations []string) string {
	bf := BatchFile{Operations: operations}
	bytes, err := json.Marshal(bf)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func getAnchorFileBytes(batchFileHash string, uniqueSuffixes []string) string {
	af := AnchorFile{
		BatchFileHash:  batchFileHash,
		UniqueSuffixes: uniqueSuffixes,
	}
	s, err := json.Marshal(af)
	if err != nil {
		panic(err)
	}
	return string(s)
}

// RegisterSteps registers sidetree txn steps
func (t *SidetreeTxnSteps) RegisterSteps(s *godog.Suite) {
	s.BeforeScenario(t.BDDContext.BeforeScenario)
	s.AfterScenario(t.BDDContext.AfterScenario)
	s.Step(`^client writes content "([^"]*)" using "([^"]*)" and the "([^"]*)" collection on the "([^"]*)" channel$`, t.writeContent)
	s.Step(`^client verifies that written content at the returned address from "([^"]*)" and the "([^"]*)" collection matches original content on the "([^"]*)" channel$`, t.readContent)
	s.Step(`^client writes operations batch file and anchor file for ID "([^"]*)" using "([^"]*)" on the "([^"]*)" channel$`, t.anchorBatch)
	s.Step(`^client creates document with ID "([^"]*)" using "([^"]*)" on the "([^"]*)" channel$`, t.createDocument)
	s.Step(`^client updates document with ID "([^"]*)" using "([^"]*)" on the "([^"]*)" channel$`, t.updateDocument)
	s.Step(`^client verifies that query by index ID "([^"]*)" from "([^"]*)" will return "([^"]*)" versions of the document on the "([^"]*)" channel$`, t.queryDocumentByIndex)
	s.Step(`^client verifies that query by index ID "([^"]*)" from "([^"]*)" will return "([^"]*)" versions of the document on the "([^"]*)" channel on peers "([^"]*)"$`, t.queryDocumentByIndexOnTargets)
}
