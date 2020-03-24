/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/docvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"
)

const (
	jsonPatchBasePath = "/fileIndex/mappings/"
)

// Validator validates the file index Sidetree document
type Validator struct {
	*docvalidator.Validator
}

// NewValidator returns a new file index document validator
func NewValidator(store docvalidator.OperationStoreClient) *Validator {
	return &Validator{
		Validator: docvalidator.New(store),
	}
}

// IsValidOriginalDocument verifies that the given payload is a valid Sidetree specific document that can be accepted by the Sidetree create operation.
func (v *Validator) IsValidOriginalDocument(payload []byte) error {
	logger.Debugf("Validating file handler original document %s", payload)

	if err := v.Validator.IsValidOriginalDocument(payload); err != nil {
		return err
	}

	fileIndexDoc := &FileIndexDoc{}
	err := jsonUnmarshal(payload, fileIndexDoc)
	if err != nil {
		return err
	}

	if fileIndexDoc.FileIndex.BasePath == "" {
		return errors.New("missing base path")
	}

	for name, id := range fileIndexDoc.FileIndex.Mappings {
		if name == "" {
			return errors.New("missing file name in mapping")
		}
		if id == "" {
			return errors.Errorf("missing ID for file name [%s]", name)
		}
	}

	return nil
}

// IsValidPayload verifies that the given payload is a valid Sidetree specific payload
// that can be accepted by the Sidetree update operations
func (v *Validator) IsValidPayload(payload []byte) error {
	logger.Debugf("Validating file handler payload %s", payload)

	if err := v.Validator.IsValidPayload(payload); err != nil {
		return err
	}

	uniqueSuffix, op, err := unmarshalUpdateOperation(payload)
	if err != nil {
		return err
	}

	for _, patch := range op.DocumentPatch {
		if err := validatePatch(patch); err != nil {
			logger.Infof("Invalid JSON patch operation for [%s]: %s", uniqueSuffix, err)
			return errors.WithMessage(err, "invalid JSON patch")
		}
	}

	return nil
}

// TransformDocument takes internal representation of document and transforms it to required representation
func (v *Validator) TransformDocument(document document.Document) (document.Document, error) {
	return document, nil
}

type patchOperation = map[string]*json.RawMessage

func validatePatch(op patchOperation) error {
	pathMsg, ok := op["path"]
	if !ok {
		return errors.New("path not found")
	}

	var path string
	if err := jsonUnmarshal(*pathMsg, &path); err != nil {
		return errors.New("invalid path")
	}

	logger.Debugf("Got path from JSON patch: [%s]", path)

	if !strings.HasPrefix(path, jsonPatchBasePath) {
		return errors.New("only the mappings section of a file index document may be modified")
	}

	return nil
}

var unmarshalUpdateOperation = func(reqPayload []byte) (string, *model.UpdateOperationData, error) {
	req := &model.UpdateRequest{}
	if err := json.Unmarshal(reqPayload, req); err != nil {
		logger.Infof("Error unmarshalling update request: %s", err)
		return "", nil, errors.New("invalid update request")
	}

	opDataBytes, err := docutil.DecodeString(req.OperationData)
	if err != nil {
		logger.Infof("Error decoding operation data for [%s]: %s", req.DidUniqueSuffix, err)
		return req.DidUniqueSuffix, nil, errors.New("invalid operation data")
	}

	logger.Debugf("Validating operation data for [%s]: %s", req.DidUniqueSuffix, opDataBytes)

	op := &model.UpdateOperationData{}
	if err := json.Unmarshal(opDataBytes, op); err != nil {
		logger.Infof("Error unmarshalling operation data for [%s]: %s", req.DidUniqueSuffix, err)
		return req.DidUniqueSuffix, nil, errors.New("invalid operation data")
	}

	return req.DidUniqueSuffix, op, nil
}

var jsonUnmarshal = func(bytes []byte, obj interface{}) error {
	return json.Unmarshal(bytes, obj)
}
