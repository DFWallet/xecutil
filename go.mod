module github.com/DFWallet/xecutil

go 1.12


replace (
	//github.com/DFWallet/xec => github.com/DFWallet/xec v0.0.5
	//github.com/DFWallet/xecutil => github.com/DFWallet/xecutil v0.0.4
)

require (
	github.com/OpenBazaar/jsonpb v0.0.0-20171123000858-37d32ddf4eef
	github.com/aead/siphash v1.0.1
	github.com/davecgh/go-spew v1.1.1
	//github.com/DFWallet/xec v0.17.2-0.20201218180520-5708823e0e99
	//github.com/DFWallet/xecutil v0.0.0-20210113190856-6ea28dff4000
	github.com/DFWallet/xec v0.0.5
	github.com/DFWallet/xecutil v0.0.4
	github.com/golang/protobuf v1.5.2
	github.com/kkdai/bstream v1.0.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210521195947-fe42d452be8f
	google.golang.org/grpc v1.38.0
)
