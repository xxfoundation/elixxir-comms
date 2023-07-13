////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// Server implementation, interface & starter function

package server

import (
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/primitives/id"
	"runtime/debug"
)

// Comms object bundles low-level connect.ProtoComms,
// and the endpoint Handler interface.
type Comms struct {
	*connect.ProtoComms
	handler Handler
	*pb.UnimplementedRemoteSyncServer
	*messages.UnimplementedGenericServer
}

// Handler describes the endpoint callbacks for remote sync.
type Handler interface {
	Read(*pb.RSReadRequest) (*pb.RSReadResponse, error)
	Write(*pb.RSWriteRequest) (*pb.RSWriteResponse, error)
	GetLastModified(*pb.RSReadRequest) (*pb.RSTimestampResponse, error)
	GetLastWrite(*messages.Ack) (*pb.RSTimestampResponse, error)
	ReadDir(*pb.RSReadRequest) (*pb.RSReadDirResponse, error)
}

// StartRemoteSync starts a new RemoteSync server on the address:port specified by localServer
// and a callback interface for remote sync operations
// with given path to public and private key for TLS connection.
func StartRemoteSync(id *id.ID, localServer string, handler Handler,
	certPem, keyPem []byte) *Comms {

	// Initialize the low-level comms listeners
	pc, err := connect.StartCommServer(id, localServer,
		certPem, keyPem, nil)
	if err != nil {
		jww.FATAL.Panicf("Unable to StartCommServer: %+v", err)
	}
	rsServer := Comms{
		handler:    handler,
		ProtoComms: pc,
	}

	// Register the high-level comms endpoint functionality
	grpcServer := rsServer.GetServer()
	pb.RegisterRemoteSyncServer(grpcServer, &rsServer)
	messages.RegisterGenericServer(grpcServer, &rsServer)

	pc.ServeWithWeb()
	return &rsServer
}

// implementationFunctions for the Handler interface.
type implementationFunctions struct {
	Read            func(*pb.RSReadRequest) (*pb.RSReadResponse, error)
	Write           func(*pb.RSWriteRequest) (*pb.RSWriteResponse, error)
	GetLastModified func(*pb.RSReadRequest) (*pb.RSTimestampResponse, error)
	GetLastWrite    func(*messages.Ack) (*pb.RSTimestampResponse, error)
	ReadDir         func(*pb.RSReadRequest) (*pb.RSReadDirResponse, error)
}

// Implementation allows users of the client library to set the
// functions that implement the node functions.
type Implementation struct {
	Functions implementationFunctions
}

// NewImplementation creates and returns a new Handler interface for implementing endpoint callbacks.
func NewImplementation() *Implementation {
	um := "UNIMPLEMENTED FUNCTION!"
	warn := func(msg string) {
		jww.WARN.Printf(msg)
		jww.WARN.Printf("%s", debug.Stack())
	}
	return &Implementation{
		Functions: implementationFunctions{
			Read: func(*pb.RSReadRequest) (*pb.RSReadResponse, error) {
				warn(um)
				return new(pb.RSReadResponse), nil
			},
			Write: func(*pb.RSWriteRequest) (*pb.RSWriteResponse, error) {
				warn(um)
				return new(pb.RSWriteResponse), nil
			},
			GetLastModified: func(*pb.RSReadRequest) (*pb.RSTimestampResponse, error) {
				warn(um)
				return new(pb.RSTimestampResponse), nil
			},
			GetLastWrite: func(*messages.Ack) (*pb.RSTimestampResponse, error) {
				warn(um)
				return new(pb.RSTimestampResponse), nil
			},
			ReadDir: func(*pb.RSReadRequest) (*pb.RSReadDirResponse, error) {
				warn(um)
				return new(pb.RSReadDirResponse), nil
			},
		},
	}
}

func (s *Implementation) Read(message *pb.RSReadRequest) (*pb.RSReadResponse, error) {
	return s.Functions.Read(message)
}
func (s *Implementation) Write(message *pb.RSWriteRequest) (*pb.RSWriteResponse, error) {
	return s.Functions.Write(message)
}
func (s *Implementation) GetLastModified(message *pb.RSReadRequest) (*pb.RSTimestampResponse, error) {
	return s.Functions.GetLastModified(message)
}
func (s *Implementation) GetLastWrite(message *messages.Ack) (*pb.RSTimestampResponse, error) {
	return s.Functions.GetLastWrite(message)
}
func (s *Implementation) ReadDir(message *pb.RSReadRequest) (*pb.RSReadDirResponse, error) {
	return s.Functions.ReadDir(message)
}
