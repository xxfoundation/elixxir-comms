////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package gateway

import (
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/privategrity/comms/mixmessages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"math"
	"net"
	"time"
)

// Passed into StartGateway to serve as an interface
// for interacting with the gateway repo
var gatewayHandler Handler

// gateway object
type gateway struct {
	gs *grpc.Server
}

// ShutDown stops the server
func (s *gateway) ShutDown() {
	s.gs.GracefulStop()
	time.Sleep(time.Millisecond * 500)
}

// Starts a new gateway on the address:port specified by localServer
// with given path to public and private key for TLS connection
func StartGateway(localServer string, handler Handler,
	certPath string, keyPath string) func() {
	// Set the gatewayHandler
	gatewayHandler = handler

	// Listen on the given address
	lis, err := net.Listen("tcp", localServer)
	if err != nil {
		jww.FATAL.Panicf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		jww.FATAL.Panicf("could not load TLS keys: %s", err)
	}

	// Create the GRPC server
	grpcServer := grpc.NewServer(grpc.Creds(creds),
		grpc.MaxConcurrentStreams(math.MaxUint32),
		grpc.MaxRecvMsgSize(33554432)) // 32 MiB
	gatewayServer := gateway{gs: grpcServer}

	go func() {
		//Make the port close when the gateway dies
		defer func() {
			err := lis.Close()
			if err != nil {
				jww.WARN.Printf("Unable to close listening port: %s", err.Error())
			}
		}()

		pb.RegisterMixMessageGatewayServer(gatewayServer.gs, &gatewayServer)

		// Register reflection service on gRPC server.
		// This blocks for the lifetime of the listener.
		reflection.Register(gatewayServer.gs)
		if err := gatewayServer.gs.Serve(lis); err != nil {
			jww.FATAL.Panicf("failed to serve: %v", err)
		}
	}()

	return gatewayServer.ShutDown
}
