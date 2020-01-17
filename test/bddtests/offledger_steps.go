/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"crypto"
	"encoding/base64"

	"github.com/DATA-DOG/godog"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

// OffLedgerSteps ...
type OffLedgerSteps struct {
	BDDContext *bddtests.BDDContext
	content    string
	address    string
}

// NewOffLedgerSteps ...
func NewOffLedgerSteps(context *bddtests.BDDContext) *OffLedgerSteps {
	return &OffLedgerSteps{BDDContext: context}
}

// DefineOffLedgerCollectionConfig defines a new off-ledger data collection configuration
func (d *OffLedgerSteps) DefineOffLedgerCollectionConfig(id, name, policy string, requiredPeerCount, maxPeerCount int32, timeToLive string) {
	d.BDDContext.DefineCollectionConfig(id,
		func(channelID string) (*pb.CollectionConfig, error) {
			sigPolicy, err := d.newChaincodePolicy(policy, channelID)
			if err != nil {
				return nil, errors.Wrapf(err, "error creating collection policy for collection [%s]", name)
			}
			return newOffLedgerCollectionConfig(name, requiredPeerCount, maxPeerCount, timeToLive, sigPolicy), nil
		},
	)
}

// DefineDCASCollectionConfig defines a new DCAS collection configuration
func (d *OffLedgerSteps) DefineDCASCollectionConfig(id, name, policy string, requiredPeerCount, maxPeerCount int32, timeToLive string) {
	d.BDDContext.DefineCollectionConfig(id,
		func(channelID string) (*pb.CollectionConfig, error) {
			sigPolicy, err := d.newChaincodePolicy(policy, channelID)
			if err != nil {
				return nil, errors.Wrapf(err, "error creating collection policy for collection [%s]", name)
			}
			return newDCASCollectionConfig(name, requiredPeerCount, maxPeerCount, timeToLive, sigPolicy), nil
		},
	)
}

func (d *OffLedgerSteps) setCASVariable(varName, value string) error {
	casKey := GetCASKey([]byte(value))
	bddtests.SetVar(varName, casKey)
	return nil
}

// getHash will compute the hash for the supplied bytes using SHA256
func getHash(bytes []byte) []byte {
	h := crypto.SHA256.New()
	// added no lint directive because there's no error from source code
	// error cannot be produced, checked google source
	h.Write(bytes) //nolint
	return h.Sum(nil)
}

// GetCASKey returns the content-addressable key for the given content
// (sha256 hash + base64 URL encoding).
func GetCASKey(content []byte) string {
	hash := getHash(content)
	buf := make([]byte, base64.URLEncoding.EncodedLen(len(hash)))
	base64.URLEncoding.Encode(buf, hash)
	return string(buf)
}

func (d *OffLedgerSteps) defineOffLedgerCollectionConfig(id, collection, policy string, requiredPeerCount int, maxPeerCount int, timeToLive string) error {
	d.DefineOffLedgerCollectionConfig(id, collection, policy, int32(requiredPeerCount), int32(maxPeerCount), timeToLive)
	return nil
}

func (d *OffLedgerSteps) defineDCASCollectionConfig(id, collection, policy string, requiredPeerCount int, maxPeerCount int, timeToLive string) error {
	d.DefineDCASCollectionConfig(id, collection, policy, int32(requiredPeerCount), int32(maxPeerCount), timeToLive)
	return nil
}

func (d *OffLedgerSteps) newChaincodePolicy(ccPolicy, channelID string) (*common.SignaturePolicyEnvelope, error) {
	return bddtests.NewChaincodePolicy(d.BDDContext, ccPolicy, channelID)
}

func newOffLedgerCollectionConfig(collName string, requiredPeerCount, maxPeerCount int32, timeToLive string, policy *common.SignaturePolicyEnvelope) *pb.CollectionConfig {
	return &pb.CollectionConfig{
		Payload: &pb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &pb.StaticCollectionConfig{
				Name:              collName,
				Type:              pb.CollectionType_COL_OFFLEDGER,
				RequiredPeerCount: requiredPeerCount,
				MaximumPeerCount:  maxPeerCount,
				TimeToLive:        timeToLive,
				MemberOrgsPolicy: &pb.CollectionPolicyConfig{
					Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
						SignaturePolicy: policy,
					},
				},
			},
		},
	}
}

func newDCASCollectionConfig(collName string, requiredPeerCount, maxPeerCount int32, timeToLive string, policy *common.SignaturePolicyEnvelope) *pb.CollectionConfig {
	return &pb.CollectionConfig{
		Payload: &pb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &pb.StaticCollectionConfig{
				Name:              collName,
				Type:              pb.CollectionType_COL_DCAS,
				RequiredPeerCount: requiredPeerCount,
				MaximumPeerCount:  maxPeerCount,
				TimeToLive:        timeToLive,
				MemberOrgsPolicy: &pb.CollectionPolicyConfig{
					Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
						SignaturePolicy: policy,
					},
				},
			},
		},
	}
}

// RegisterSteps registers off-ledger steps
func (d *OffLedgerSteps) RegisterSteps(s *godog.Suite) {
	s.BeforeScenario(d.BDDContext.BeforeScenario)
	s.AfterScenario(d.BDDContext.AfterScenario)
	s.Step(`^variable "([^"]*)" is assigned the CAS key of value "([^"]*)"$`, d.setCASVariable)
	s.Step(`^off-ledger collection config "([^"]*)" is defined for collection "([^"]*)" as policy="([^"]*)", requiredPeerCount=(\d+), maxPeerCount=(\d+), and timeToLive=([^"]*)$`, d.defineOffLedgerCollectionConfig)
	s.Step(`^DCAS collection config "([^"]*)" is defined for collection "([^"]*)" as policy="([^"]*)", requiredPeerCount=(\d+), maxPeerCount=(\d+), and timeToLive=([^"]*)$`, d.defineDCASCollectionConfig)

}
