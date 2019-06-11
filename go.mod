module github.com/trustbloc/sidetree-fabric

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/loads v0.19.0
	github.com/go-openapi/runtime v0.19.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190328182020-93c3fcb272be
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.3.0
	github.com/trustbloc/sidetree-core-go v0.0.0-20190604193932-b3a21a189580
	github.com/trustbloc/sidetree-node v0.0.0-20190605161025-0df7c418272b
)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos => github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos v0.0.0-20190328182020-93c3fcb272be
