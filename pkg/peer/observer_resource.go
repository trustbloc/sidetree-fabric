/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
)

type observerResource struct {
	observer *observer.Observer
}

func newObserver(providers *observer.Providers) *observerResource {
	logger.Infof("Creating Sidetree observer")

	return &observerResource{
		observer: observer.New(providers),
	}
}

// ChannelJoined is invoked when the peer joins a channel
func (r *observerResource) ChannelJoined(channelID string) {
	logger.Infof("Joined channel [%s]", channelID)

	if err := r.observer.Start(channelID); err != nil {
		logger.Errorf("Error starting observer for channel [%s]", channelID)
	}
}

// Close stops the observer
func (r *observerResource) Close() {
	logger.Infof("Stopping observer")

	r.observer.Stop()
}
