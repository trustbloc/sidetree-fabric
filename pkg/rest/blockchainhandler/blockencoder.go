/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/pkg/errors"
)

type blockEncoder struct {
	encoding DataEncoding
}

func newBlockEncoder(encoding DataEncoding) *blockEncoder {
	return &blockEncoder{encoding: encoding}
}

func (e *blockEncoder) encode(block *cb.Block) (Block, error) {
	switch e.encoding {
	case DataEncodingJSON:
		return e.encodeJSON(block)
	case DataEncodingBase64:
		return e.encodeBase64(block, base64.StdEncoding)
	case DataEncodingBase64URL:
		return e.encodeBase64(block, base64.URLEncoding)
	default:
		return nil, errors.Errorf("unsupported data encoding %s", e.encoding)
	}
}

func (e *blockEncoder) encodeJSON(block *cb.Block) (Block, error) {
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, block); err != nil {
		return nil, err
	}

	var jsonBlock Block
	if err := json.Unmarshal(buf.Bytes(), &jsonBlock); err != nil {
		return nil, err
	}

	return jsonBlock, nil
}

func (e *blockEncoder) encodeBase64(block *cb.Block, encoding *base64.Encoding) (Block, error) {
	jsonBlock := make(Block)

	header := make(map[string]interface{})
	header[numberField] = block.Header.Number
	header[dataHashField] = encoding.EncodeToString(block.Header.DataHash)
	header[previousHashField] = encoding.EncodeToString(block.Header.PreviousHash)

	jsonBlock[headerField] = header
	jsonBlock[dataField] = encoding.EncodeToString(bytes.Join(block.Data.Data, nil))
	jsonBlock[metadataField] = encoding.EncodeToString(bytes.Join(block.Metadata.Metadata, nil))

	return jsonBlock, nil
}

func (h *handler) dataEncoding(req *http.Request) DataEncoding {
	encoding := getDataEncoding(req)
	if encoding == "" {
		encoding = DataEncodingJSON
	}

	return encoding
}
