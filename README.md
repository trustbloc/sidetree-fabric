[![Release](https://img.shields.io/github/release/trustbloc/sidetree-fabric.svg?style=flat-square)](https://github.com/trustbloc/sidetree-fabric/releases/latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/sidetree-fabric/master/LICENSE)

[![Build Status](https://dev.azure.com/trustbloc/sidetree/_apis/build/status/trustbloc.sidetree-fabric?branchName=master)](https://dev.azure.com/trustbloc/sidetree/_build/latest?definitionId=13&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/trustbloc/sidetree-fabric?style=flat-square)](https://goreportcard.com/report/github.com/trustbloc/sidetree-fabric)
[![codebeat badge](https://codebeat.co/badges/d549a1a4-372c-416b-ae56-7b6e395b3a56)](https://codebeat.co/projects/github-com-trustbloc-sidetree-fabric-master)
[![codecov](https://codecov.io/gh/trustbloc/sidetree-fabric/branch/master/graph/badge.svg)](https://codecov.io/gh/trustbloc/sidetree-fabric)


# sidetree-fabric

## Build

The project is built using make. 

BDD test suit can be run via `make bddtests`

## Run

To run a Sidetree node along with Hyperledger Fabric you can use docker-compose.

First run the compose itself via

1. `cd test/bddtests/fixtures/`
2. `docker-compose up –force-recreate`
This will start up the node and Fabric but you need to set up the ledger first. 
This is done by running BDD tests outside of make (after the containers have been started):

3. `cd test/bddtests`
4. `DISABLE_COMPOSITION=true go test`

After that you can invoke the Sidetree REST API at the following URL: http://localhost:48326/document


To bring everything down run `docker-compose down`

## Contributing
Thank you for your interest in contributing. Please see our [community contribution guidelines](https://github.com/trustbloc/community/blob/master/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
