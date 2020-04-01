/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/monitor"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type monitorController struct {
	channelID string
	monitor   *monitor.Monitor
}

func newMonitorController(channelID string, peerConfig peerConfig, monitorCfg config.Monitor, providers *monitor.ClientProviders, opStoreProvider common.OperationStoreProvider) *monitorController {
	var m *monitor.Monitor
	if role.IsMonitor() {
		m = monitor.New(channelID, peerConfig.PeerID(), monitorCfg.Period, providers, opStoreProvider)
	}

	return &monitorController{
		channelID: channelID,
		monitor:   m,
	}
}

// Start starts the Sidetree monitor if it is set
func (m *monitorController) Start() error {
	if m.monitor != nil {
		logger.Debugf("[%s] Starting Sidetree monitor ...", m.channelID)

		return m.monitor.Start()
	}

	return nil
}

// Stop stops the Sidetree monitor if it is set
func (m *monitorController) Stop() {
	if m.monitor != nil {
		logger.Debugf("[%s] Stopping Sidetree monitor ...", m.channelID)

		m.monitor.Stop()
	}
}
