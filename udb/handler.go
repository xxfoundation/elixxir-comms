////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// Contains callback interface for registration functionality

package udb

import (
	//	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
	//	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/primitives/id"
	//	"google.golang.org/grpc/reflection"
	"runtime/debug"
)

// Registration object used to implement
// endpoints and top-level comms functionality
type Comms struct {
	*connect.ProtoComms
	handler Handler // an object that implements the interface below, which
	// has all the functions called by endpoint.go
	*pb.UnimplementedUDBServer
	*messages.UnimplementedGenericServer
}

// StartServer starts a new server on the address:port specified by localServer
// and a callback interface for server operations
// with given path to public and private key for TLS connection
func StartServer(id *id.ID, localServer string, handler Handler,
	certPEMblock, keyPEMblock []byte) *Comms {
	pc, err := connect.StartCommServer(id, localServer,
		certPEMblock, keyPEMblock, nil)
	if err != nil {
		jww.FATAL.Panicf("Unable to start comms server: %+v", err)
	}

	udbServer := Comms{
		ProtoComms: pc,
		handler:    handler,
	}
	pb.RegisterUDBServer(udbServer.GetServer(), &udbServer)
	messages.RegisterGenericServer(udbServer.GetServer(), &udbServer)

	pc.ServeWithWeb()
	return &udbServer
}

// Handler is the interface udb has to implement to integrate with the comms
// library properly.
type Handler interface {
	// RegisterUser handles registering a user into the database
	RegisterUser(registration *pb.UDBUserRegistration) (*messages.Ack, error)
	// RemoveUser deletes this user registration and blocks anyone from ever
	// registering under that username again.
	// The fact removal request must be for the username or it will not work.
	RemoveUser(request *pb.FactRemovalRequest) (*messages.Ack, error)
	// RegisterFact handles registering a fact into the database
	RegisterFact(msg *pb.FactRegisterRequest) (*pb.FactRegisterResponse, error)
	// ConfirmFact checks a Fact against the Fact database
	ConfirmFact(msg *pb.FactConfirmRequest) (*messages.Ack, error)
	// RemoveFact deletes a fact from its associated ID.
	// You cannot RemoveFact on a username. Callers must RemoveUser and reregister.
	RemoveFact(request *pb.FactRemovalRequest) (*messages.Ack, error)
	// RequestChannelLease requests a signature & lease on a user's ed25519 public key from user discovery for use in channels
	RequestChannelLease(request *pb.ChannelLeaseRequest) (*pb.ChannelLeaseResponse, error)
	// ValidateUsername validates that a user owns a username by signing the contents of the
	// mixmessages.UsernameValidationRequest.
	ValidateUsername(request *pb.UsernameValidationRequest) (*pb.UsernameValidation, error)
}

// implementationFunctions are the actual implementations of
type implementationFunctions struct {
	// This is the function "implementation" -- inside UDB we will
	// set this to be the UDB version of the function. By default
	// it's a dummy function that returns nothing (see NewImplementation
	// below).

	// RegisterUser handles registering a user into the database
	RegisterUser func(registration *pb.UDBUserRegistration) (*messages.Ack, error)
	// RemoveUser deletes this user registration and blocks anyone from ever
	// registering under that username again.
	// The fact removal request must be for the username or it will not work.
	RemoveUser func(request *pb.FactRemovalRequest) (*messages.Ack, error)
	// RegisterFact handles registering a fact into the database
	RegisterFact func(request *pb.FactRegisterRequest) (*pb.FactRegisterResponse, error)
	// ConfirmFact checks a Fact against the Fact database
	ConfirmFact func(request *pb.FactConfirmRequest) (*messages.Ack, error)
	// RemoveFact deletes a fact from its associated ID.
	// You cannot RemoveFact on a username. Callers must RemoveUser and reregister.
	RemoveFact func(request *pb.FactRemovalRequest) (*messages.Ack, error)
	// RequestChannelLease requests a signature & lease on a user's ed25519 public key from user discovery for use in channels
	RequestChannelLease func(request *pb.ChannelLeaseRequest) (*pb.ChannelLeaseResponse, error)
	// ValidateUsername validates that a user owns a username by signing the contents of the
	// mixmessages.UsernameValidationRequest.
	ValidateUsername func(request *pb.UsernameValidationRequest) (*pb.UsernameValidation, error)
}

// Implementation allows users of the client library to set the
// functions that implement the node functions
type Implementation struct {
	Functions implementationFunctions
}

// NewImplementation returns a Implementation struct with all of the
// function pointers returning nothing and printing an error.
// Inside UDB, you would call this, then set all functions to your
// own UDB version of the function.
func NewImplementation() *Implementation {
	um := "UNIMPLEMENTED FUNCTION!"
	warn := func(msg string) {
		jww.WARN.Printf(msg)
		jww.WARN.Printf("%s", debug.Stack())
	}
	return &Implementation{
		Functions: implementationFunctions{
			// Stub for RegisterUser which returns a blank message and prints a warning
			RegisterUser: func(registration *pb.UDBUserRegistration) (*messages.Ack, error) {
				warn(um)
				return &messages.Ack{}, nil
			},
			// Stub for RemoveUser which returns a blank message and prints a warning
			RemoveUser: func(request *pb.FactRemovalRequest) (*messages.Ack, error) {
				warn(um)
				return &messages.Ack{}, nil
			},
			// Stub for RegisterFact which returns a blank message and prints a warning
			RegisterFact: func(request *pb.FactRegisterRequest) (*pb.FactRegisterResponse, error) {
				warn(um)
				return &pb.FactRegisterResponse{}, nil
			},
			// Stub for ConfirmFact which returns a blank message and prints a warning
			ConfirmFact: func(request *pb.FactConfirmRequest) (*messages.Ack, error) {
				warn(um)
				return &messages.Ack{}, nil
			},
			// Stub for RemoveFact which returns a blank message and prints a warning
			RemoveFact: func(request *pb.FactRemovalRequest) (*messages.Ack, error) {
				warn(um)
				return &messages.Ack{}, nil
			},
			RequestChannelLease: func(request *pb.ChannelLeaseRequest) (*pb.ChannelLeaseResponse, error) {
				warn(um)
				return &pb.ChannelLeaseResponse{}, nil
			},
			ValidateUsername: func(request *pb.UsernameValidationRequest) (*pb.UsernameValidation, error) {
				warn(um)
				return &pb.UsernameValidation{}, nil
			},
		},
	}
}

// RegisterUser is called by the RegisterUser in endpoint.go. It calls the corresponding function in the interface.
func (s *Implementation) RegisterUser(registration *pb.UDBUserRegistration) (*messages.Ack, error) {
	return s.Functions.RegisterUser(registration)
}

// RemoveUser is called by the RemoveUser in endpoint.go. It calls the corresponding function in the interface.
func (s *Implementation) RemoveUser(request *pb.FactRemovalRequest) (*messages.Ack, error) {
	return s.Functions.RemoveUser(request)
}

// RegisterFact is called by the RegisterFact in endpoint.go. It calls the corresponding function in the interface.
func (s *Implementation) RegisterFact(request *pb.FactRegisterRequest) (*pb.FactRegisterResponse, error) {
	return s.Functions.RegisterFact(request)
}

// ConfirmFact is called by the ConfirmFact in endpoint.go. It calls the corresponding function in the interface.
func (s *Implementation) ConfirmFact(request *pb.FactConfirmRequest) (*messages.Ack, error) {
	return s.Functions.ConfirmFact(request)
}

// RemoveFact is called by the RemoveFact in endpoint.go. It calls the corresponding function in the interface.
func (s *Implementation) RemoveFact(request *pb.FactRemovalRequest) (*messages.Ack, error) {
	return s.Functions.RemoveFact(request)
}

// RequestChannelLease is called by the RequestChannelAuthentication in endpoint.go.  It calls the corresponding function in the interface
func (s *Implementation) RequestChannelLease(request *pb.ChannelLeaseRequest) (*pb.ChannelLeaseResponse, error) {
	return s.Functions.RequestChannelLease(request)
}

// ValidateUsername validates that a user owns a username by signing the contents of the
// mixmessages.UsernameValidationRequest.
func (s *Implementation) ValidateUsername(request *pb.UsernameValidationRequest) (*pb.UsernameValidation, error) {
	return s.Functions.ValidateUsername(request)
}
