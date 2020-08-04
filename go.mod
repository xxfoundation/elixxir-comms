module gitlab.com/elixxir/comms

go 1.13

require (
	github.com/golang/protobuf v1.4.2
	github.com/pkg/errors v0.9.1
	github.com/spf13/jwalterweatherman v1.1.0
	gitlab.com/elixxir/crypto v0.0.0-20200804182833-984246dea2c4
	gitlab.com/elixxir/primitives v0.0.0-20200804182913-788f47bded40
	gitlab.com/xx_network/comms v0.0.0-20200804220700-a5a9bd64204a
	gitlab.com/xx_network/primitives v0.0.0-20200804183002-f99f7a7284da
	gitlab.com/xx_network/ring v0.0.2
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/grpc v1.30.0
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.27.1
