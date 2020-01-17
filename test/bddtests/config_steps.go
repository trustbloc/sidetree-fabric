/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

type configSteps struct {
	BDDContext *bddtests.BDDContext
}

// NewConfigSteps returns ledger config steps
func NewConfigSteps(context *bddtests.BDDContext) *configSteps {
	return &configSteps{BDDContext: context}
}

// loadConfigFromFile loads configuration data from a file and sets the value to the given variable
func (d *configSteps) loadConfigFromFile(varName, filePath string) error {
	logger.Infof("Loading config from file [%s] to variable [%s]", filePath, varName)

	cfgBytes, err := readFile(filePath)
	if err != nil {
		return err
	}

	cfg := &Config{}
	if err := json.Unmarshal(cfgBytes, cfg); err != nil {
		return err
	}

	// Replace all of the file references with actual config
	newCfg, err := newConfigPreProcessor(filePath).preProcess(cfg)
	if err != nil {
		return err
	}

	cfgBytes, err = json.Marshal(newCfg)
	if err != nil {
		return err
	}

	bddtests.SetVar(varName, string(cfgBytes))

	return nil
}

// RegisterSteps registers config steps
func (d *configSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^variable "([^"]*)" is assigned config from file "([^"]*)"$`, d.loadConfigFromFile)
}

type configPreprocessor struct {
	basePath string
}

func newConfigPreProcessor(basePath string) *configPreprocessor {
	return &configPreprocessor{basePath: basePath}
}

func (cp *configPreprocessor) preProcess(cfg *Config) (*Config, error) {
	peers, err := cp.visitPeers(cfg.Peers)
	if err != nil {
		return nil, err
	}
	apps, err := cp.visitApps(cfg.Apps)
	if err != nil {
		return nil, err
	}
	return &Config{
		MspID: cfg.MspID,
		Peers: peers,
		Apps:  apps,
	}, nil
}

func (cp *configPreprocessor) visitPeers(srcPeers []*Peer) ([]*Peer, error) {
	peers := make([]*Peer, len(srcPeers))
	for i, p := range srcPeers {
		apps, err := cp.visitApps(p.Apps)
		if err != nil {
			return nil, err
		}
		peers[i] = &Peer{
			PeerID: p.PeerID,
			Apps:   apps,
		}
	}
	return peers, nil
}

func (cp *configPreprocessor) visitApps(srcApps []*App) ([]*App, error) {
	apps := make([]*App, len(srcApps))
	for i, a := range srcApps {
		var config string
		var components []*Component
		var err error
		if a.Config != "" {
			config, err = cp.visitConfigString(a.Config)
			if err != nil {
				return nil, err
			}
		} else {
			components, err = cp.visitComponents(a.Components)
			if err != nil {
				return nil, err
			}
		}
		apps[i] = &App{
			AppName:    a.AppName,
			Version:    a.Version,
			Format:     a.Format,
			Config:     config,
			Components: components,
		}
	}
	return apps, nil
}

func (cp *configPreprocessor) visitComponents(srcComponents []*Component) ([]*Component, error) {
	components := make([]*Component, len(srcComponents))
	for i, c := range srcComponents {
		config, err := cp.visitConfigString(c.Config)
		if err != nil {
			return nil, err
		}
		components[i] = &Component{
			Name:    c.Name,
			Version: c.Version,
			Format:  c.Format,
			Config:  config,
		}
	}
	return components, nil
}

func (cp *configPreprocessor) visitConfigString(srcConfig string) (string, error) {
	// Substitute all of the file refs with the actual contents of the file
	if !strings.HasPrefix(srcConfig, "file://") {
		return srcConfig, nil
	}

	refFilePath := srcConfig[7:]
	contents, err := cp.readFileRef(refFilePath)
	if err != nil {
		return "", errors.Wrapf(err, "error retrieving contents of file [%s]", refFilePath)
	}
	return string(contents), nil
}

func (cp *configPreprocessor) readFileRef(refPath string) ([]byte, error) {
	var path string
	if filepath.IsAbs(refPath) || cp.basePath == "" {
		path = refPath
	} else {
		// The path is relative to the source config file
		path = filepath.Join(filepath.Dir(cp.basePath), refPath)
	}
	return readFile(path)
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, errors.WithMessagef(err, "error opening file [%s]", path)
	}
	defer func() {
		if e := file.Close(); e != nil {
			// This shouldn't happen
			panic(err.Error())
		}
	}()

	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.WithMessagef(err, "error reading config file [%s]", path)
	}
	return configBytes, nil
}

// Format specifies the format of the configuration
type Format string

// Config contains zero or more application configurations and zero or more peer-specific application configurations
type Config struct {
	// MspID is the ID of the MSP
	MspID string
	// Peers contains configuration for zero or more peers
	Peers []*Peer `json:",omitempty"`
	// Apps contains configuration for zero or more application
	Apps []*App `json:",omitempty"`
}

// Peer contains a collection of application configurations for a given peer
type Peer struct {
	// PeerID is the unique ID of the peer
	PeerID string
	// Apps contains configuration for one or more application
	Apps []*App
}

// App contains the configuration for an application and/or multiple sub-components.
type App struct {
	// Name is the name of the application
	AppName string
	// Version is the version of the config
	Version string
	// Format describes the format of the data
	Format Format
	// Config contains the actual configuration
	Config string
	// Components zero or more component configs
	Components []*Component `json:",omitempty"`
}

// Component contains the configuration for an application component.
type Component struct {
	// Name is the name of the component
	Name string
	// Version is the version of the config
	Version string
	// Format describes the format of the data
	Format Format
	// Config contains the actual configuration
	Config string
}
