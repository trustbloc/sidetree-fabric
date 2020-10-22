/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validator

import (
	"encoding/json"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/docvalidator/docvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/model"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	jsonPatchBasePath = "/fileIndex/mappings/"
)

// FileIdxValidator validates the file index Sidetree document
type FileIdxValidator struct {
	*docvalidator.Validator
}

// NewFileIdxValidator returns a new file index document validator
func NewFileIdxValidator(store docvalidator.OperationStoreClient) *FileIdxValidator {
	return &FileIdxValidator{
		Validator: docvalidator.New(store),
	}
}

// IsValidOriginalDocument verifies that the given payload is a valid Sidetree specific document that can be accepted by the Sidetree create operation.
func (v *FileIdxValidator) IsValidOriginalDocument(payload []byte) error {
	logger.Debugf("Validating file handler original document %s", payload)

	if err := v.Validator.IsValidOriginalDocument(payload); err != nil {
		return err
	}

	fileIndexDoc := &filehandler.FileIndexDoc{}
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
func (v *FileIdxValidator) IsValidPayload(payload []byte) error {
	logger.Debugf("Validating file handler payload %s", payload)

	if err := v.Validator.IsValidPayload(payload); err != nil {
		return err
	}

	uniqueSuffix, op, err := unmarshalUpdateOperation(payload)
	if err != nil {
		return err
	}

	for _, p := range op.Patches {
		if err := validatePatch(p); err != nil {
			logger.Infof("Invalid JSON patch data for [%s]: %s", uniqueSuffix, err)
			return errors.WithMessage(err, "invalid JSON patch")
		}
	}

	return nil
}

func validatePatch(p patch.Patch) error {
	action, err := p.GetAction()
	if err != nil {
		return errors.New("unable to get patch action")
	}

	if action != patch.JSONPatch {
		return errors.Errorf("patch action '%s' not supported", action)
	}

	patches, err := p.GetValue()
	if err != nil {
		return errors.WithMessage(err, "invalid patch value")
	}

	bytes, err := json.Marshal(patches)
	if err != nil {
		return err
	}

	jsonPatches, err := jsonpatch.DecodePatch(bytes)
	if err != nil {
		return err
	}

	for _, p := range jsonPatches {
		pathMsg, ok := p["path"]
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
	}

	return nil
}

var unmarshalUpdateOperation = func(reqPayload []byte) (string, *model.DeltaModel, error) {
	req := &model.UpdateRequest{}
	if err := json.Unmarshal(reqPayload, req); err != nil {
		logger.Infof("Error unmarshalling update request: %s", err)
		return "", nil, errors.New("invalid update request")
	}

	return req.DidSuffix, req.Delta, nil
}

var jsonUnmarshal = func(bytes []byte, obj interface{}) error {
	return json.Unmarshal(bytes, obj)
}
