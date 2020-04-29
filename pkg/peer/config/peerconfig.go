/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	viper "github.com/spf13/viper2015"
)

const (
	sidetreeHostKey        = "sidetree.host"
	sidetreePortKey        = "sidetree.port"
	sidetreeTLSCertificate = "sidetree.tls.cert.file"
	sidetreeTLSKey         = "sidetree.tls.key.file"
	sidetreeAPITokens      = "sidetree.api.tokens"

	confPeerFileSystemPath = "peer.fileSystemPath"
	sidetreeOperationsDir  = "sidetree_ops"
)

// Peer holds the Sidetree peer config
type Peer struct {
	sidetreeHost           string
	sidetreePort           int
	sidetreeTLSCertificate string
	sidetreeTLSKey         string
	levelDBOpQueueBasePath string
	sidetreeAPITokens      map[string]string
}

// NewPeer returns a new peer config
func NewPeer() *Peer {
	tokens := make(map[string]string)
	if viper.GetString(sidetreeAPITokens) != "" {
		for _, field := range strings.Split(viper.GetString(sidetreeAPITokens), ":") {
			split := strings.Split(field, "=")
			switch len(split) {
			case 2:
				tokens[split[0]] = split[1]
			default:
				logger.Warnf("invalid token '%s'", field)
			}
		}
	}

	return &Peer{
		sidetreeHost:           viper.GetString(sidetreeHostKey),
		sidetreePort:           viper.GetInt(sidetreePortKey),
		sidetreeTLSCertificate: viper.GetString(sidetreeTLSCertificate),
		sidetreeTLSKey:         viper.GetString(sidetreeTLSKey),
		levelDBOpQueueBasePath: filepath.Join(filepath.Clean(viper.GetString(confPeerFileSystemPath)), sidetreeOperationsDir),
		sidetreeAPITokens:      tokens,
	}
}

// SidetreeListenURL returns the URL on which the Sidetree REST service should be started
func (c *Peer) SidetreeListenURL() (string, error) {
	host := c.sidetreeHost
	if host == "" {
		host = "0.0.0.0"
	}

	if c.sidetreePort == 0 {
		return "", errors.New("port is not set for REST service")
	}

	return fmt.Sprintf("%s:%d", host, c.sidetreePort), nil
}

// LevelDBOpQueueBasePath returns the base path of the directory to store LevelDB operation queues
func (c *Peer) LevelDBOpQueueBasePath() string {
	return c.levelDBOpQueueBasePath
}

// SidetreeTLSCertificate returns the tls certificate
func (c *Peer) SidetreeTLSCertificate() string {
	return c.sidetreeTLSCertificate
}

// SidetreeTLSKey returns the tls key
func (c *Peer) SidetreeTLSKey() string {
	return c.sidetreeTLSKey
}

// SidetreeAPIToken returns api token
func (c *Peer) SidetreeAPIToken(name string) string {
	return c.sidetreeAPITokens[name]
}
