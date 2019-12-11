/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
)

var logger = flogging.MustGetLogger("sidetree_peer")

// Initialize initializes the required resources for peer startup
func Initialize() {
	resource.Register(client.NewBlockchainProvider)
	resource.Register(newObserver)
}

type observerResource struct {
	observer *observer.Observer
}

func newObserver(providers *observer.Providers) *observerResource {
	logger.Infof("Initializing observer")

	return &observerResource{
		observer: observer.New(providers),
	}
}

// ChannelJoined is invoked when the peer joins a channel
func (r *observerResource) ChannelJoined(channelID string) {
	logger.Infof("Joined channel [%s]", channelID)
	started, err := r.observer.Start(channelID)
	if err != nil {
		logger.Errorf("Error starting observer for channel [%s]", channelID)
		return
	}

	if started {
		logger.Infof("Successfully started observer for channel [%s]", channelID)
	}
}

// Close stops the observer
func (r *observerResource) Close() {
	logger.Infof("Stopping observer")
	r.observer.Stop()
}
