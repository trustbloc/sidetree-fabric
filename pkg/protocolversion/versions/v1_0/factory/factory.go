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
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/doccomposer"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/doctransformer/didtransformer"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/doctransformer/doctransformer"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/docvalidator/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/operationapplier"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/operationparser"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/txnprocessor"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/txnprovider"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	vcommon "github.com/trustbloc/sidetree-fabric/pkg/protocolversion/versions/common"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/versions/v1_0/validator"
)

// Factory implements version 0.1 of the Sidetree protocol
type Factory struct {
}

// New returns a version 0.1 implementation of the Sidetree protocol
func New() *Factory {
	return &Factory{}
}

// Create creates a new protocol version
func (v *Factory) Create(version string, p protocol.Protocol, casClient cas.Client, opStore ctxcommon.OperationStore, docType common.DocumentType, sidetreeCfg config.Sidetree) (protocol.Version, error) {
	parser := operationparser.New(p)
	cp := compression.New(compression.WithDefaultAlgorithms())
	opp := txnprovider.NewOperationProvider(p, parser, casClient, cp)
	oh := txnprovider.NewOperationHandler(p, casClient, cp, parser)
	dc := doccomposer.New()
	oa := operationapplier.New(p, parser, dc)

	txnProcessor := txnprocessor.New(
		&txnprocessor.Providers{
			OpStore:                   opStore,
			OperationProtocolProvider: opp,
		},
	)

	dv, dt, err := createDocumentProviders(docType, opStore, sidetreeCfg)
	if err != nil {
		return nil, err
	}

	return &vcommon.ProtocolVersion{
		VersionStr:     version,
		P:              p,
		TxnProcessor:   txnProcessor,
		OpParser:       parser,
		OpApplier:      oa,
		DocComposer:    dc,
		OpHandler:      oh,
		OpProvider:     opp,
		DocValidator:   dv,
		DocTransformer: dt,
	}, nil
}

func createDocumentProviders(docType common.DocumentType, opStore ctxcommon.OperationStore, sidetreeCfg config.Sidetree) (protocol.DocumentValidator, protocol.DocumentTransformer, error) {
	switch docType {
	case common.FileIndexType:
		return validator.NewFileIdxValidator(opStore), doctransformer.New(), nil
	case common.DIDDocType:
		return didvalidator.New(opStore), didtransformer.New(didtransformer.WithMethodContext(sidetreeCfg.MethodContext), didtransformer.WithBase(sidetreeCfg.EnableBase)), nil
	default:
		return nil, nil, fmt.Errorf("unsupported document type: [%s]", docType)
	}
}
