module github.com/trustbloc/sidetree-fabric

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190328182020-93c3fcb272be
	github.com/pkg/errors v0.8.0
	github.com/stretchr/testify v1.3.0
	github.com/trustbloc/sidetree-core-go v0.0.0-20190604193932-b3a21a189580
)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos => github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos v0.0.0-20190328182020-93c3fcb272be
