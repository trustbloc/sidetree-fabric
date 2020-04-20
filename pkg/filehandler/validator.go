/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/docvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
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

	for _, p := range op.Patches {
		if err := validatePatch(p); err != nil {
			logger.Infof("Invalid JSON patch data for [%s]: %s", uniqueSuffix, err)
			return errors.WithMessage(err, "invalid JSON patch")
		}
	}

	return nil
}

// TransformDocument takes internal representation of document and transforms it to required representation
func (v *Validator) TransformDocument(doc document.Document) (*document.ResolutionResult, error) {
	resolutionResult := &document.ResolutionResult{
		Document:       doc,
		MethodMetadata: document.MethodMetadata{},
	}

	processKeys(doc, resolutionResult)

	return resolutionResult, nil
}

// generic documents will most likely only contain operation keys
// operation keys are not part of external document but resolution result
func processKeys(internal document.Document, resolutionResult *document.ResolutionResult) {
	var operationPublicKeys []document.PublicKey

	var nonOperationsKeys []document.PublicKey
	for _, pk := range internal.PublicKeys() {
		externalPK := make(document.PublicKey)
		externalPK[document.IDProperty] = internal.ID() + "#" + pk.ID()
		externalPK[document.TypeProperty] = pk.Type()
		externalPK[document.ControllerProperty] = internal[document.IDProperty]
		externalPK[document.PublicKeyJwkProperty] = pk.JWK()

		usages := pk.Usage()
		delete(pk, document.UsageProperty)

		if document.IsOperationsKey(usages) {
			operationPublicKeys = append(operationPublicKeys, externalPK)
		} else {
			nonOperationsKeys = append(nonOperationsKeys, externalPK)
		}
	}

	if len(nonOperationsKeys) > 0 {
		internal[document.PublicKeyProperty] = nonOperationsKeys
	} else {
		delete(internal, document.PublicKeyProperty)
	}

	resolutionResult.MethodMetadata.OperationPublicKeys = operationPublicKeys
}

func validatePatch(p patch.Patch) error {
	if p.GetAction() != patch.JSONPatch {
		return errors.Errorf("patch action '%s' not supported", p.GetAction())
	}

	patches := p.GetValue(patch.PatchesKey)
	if patches == "" {
		return errors.New("missing patches string value")
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

	patchDataBytes, err := docutil.DecodeString(req.Delta)
	if err != nil {
		logger.Infof("Error decoding patch data for [%s]: %s", req.DidSuffix, err)
		return req.DidSuffix, nil, errors.New("invalid patch data")
	}

	logger.Debugf("Validating patch data for [%s]: %s", req.DidSuffix, patchDataBytes)

	op := &model.DeltaModel{}
	if err := json.Unmarshal(patchDataBytes, op); err != nil {
		logger.Infof("Error unmarshalling patch data for [%s]: %s", req.DidSuffix, err)
		return req.DidSuffix, nil, errors.New("invalid patch data")
	}

	return req.DidSuffix, op, nil
}

var jsonUnmarshal = func(bytes []byte, obj interface{}) error {
	return json.Unmarshal(bytes, obj)
}
