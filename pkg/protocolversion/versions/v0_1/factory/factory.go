/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package factory

import (
	"fmt"

	"github.com/trustbloc/sidetree-core-go/pkg/api/cas"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/compression"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/doccomposer"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/docvalidator/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/operationapplier"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/operationparser"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/txnprocessor"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/txnprovider"

	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/common"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/versions/v0_1/validator"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

// Factory implements version 0.1 of the Sidetree protocol
type Factory struct {
}

// New returns a version 0.1 implementation of the Sidetree protocol
func New() *Factory {
	return &Factory{}
}

// Create creates a new protocol version
func (v *Factory) Create(p protocol.Protocol, casClient cas.Client, opStore ctxcommon.OperationStore, docType sidetreehandler.DocumentType) (protocol.Version, error) {
	parser := operationparser.New(p)
	cp := compression.New(compression.WithDefaultAlgorithms())
	opp := txnprovider.NewOperationProvider(p, parser, casClient, cp)
	oh := txnprovider.NewOperationHandler(p, casClient, cp)
	dc := doccomposer.New()
	oa := operationapplier.New(p, parser, dc)

	txnProcessor := txnprocessor.New(
		&txnprocessor.Providers{
			OpStore:                   opStore,
			OperationProtocolProvider: opp,
		},
	)

	dv, err := createDocumentValidator(docType, opStore)
	if err != nil {
		return nil, err
	}

	return &common.ProtocolVersion{
		P:            p,
		TxnProcessor: txnProcessor,
		OpParser:     parser,
		OpApplier:    oa,
		DocComposer:  dc,
		OpHandler:    oh,
		OpProvider:   opp,
		DocValidator: dv,
	}, nil
}

func createDocumentValidator(docType sidetreehandler.DocumentType, opStore ctxcommon.OperationStore) (protocol.DocumentValidator, error) {
	switch docType {
	case sidetreehandler.FileIndexType:
		return validator.NewFileIdxValidator(opStore), nil
	case sidetreehandler.DIDDocType:
		return didvalidator.New(opStore), nil
	default:
		return nil, fmt.Errorf("unsupported document type: [%s]", docType)
	}
}
